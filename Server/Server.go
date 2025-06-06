package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Map untuk menyimpan koleksi client yg aktif
var (
	clients   = make(map[net.Conn]string)
	clientsMu sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":2000") // buka listener
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server started. Waiting for clients...")

	// accept koneksi dari client
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// add client ke daftar clients aktif di protect Mutex
		// clientsMu.Lock()
		// clients[conn] = true
		// clientsMu.Unlock()

		fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())
		go handleClient(conn) // handle comms client pada goroutine terpisah
		// conn.Close() // langsung tutup koneksi, hanya untuk cek koneksi saja
	}
}

// Handle setiap client
func handleClient(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// read username dari client
	username, err := checkUsername(conn, reader)
	if err != nil {
		conn.Close()
		return
	}
	// username = strings.TrimSpace(username)

	// simpan username dari client
	clientsMu.Lock()
	clients[conn] = username
	clientsMu.Unlock()

	fmt.Printf("%s connected\n", username)
	// send notif ke client, ketika ada client baru yang connect
	formattedNotif := fmt.Sprintf("%s jumped in!\n", username)
	broadcastMessage(conn, formattedNotif)

	// defer anonymous class -> dieksekusi ketika client disconnect
	defer func() {
		clientsMu.Lock()
		delete(clients, conn) // hapus client dari daftar aktif
		clientsMu.Unlock()
		conn.Close() // tutup koneksi
		// fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
		fmt.Printf("%s disconnected\n", username)

		// send notif ke client, ketika ada client baru yang disconnect
		formattedNotif := fmt.Sprintf("%s has left!\n", username)
		broadcastMessage(conn, formattedNotif)
	}()
	// defer conn.Close()

	for {
		// read message dari client
		message, err := reader.ReadString('\n')
		if err != nil {
			// Client disconnected or error
			// fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
			// return
			break
		}
		// fmt.Printf("Received message from %s: %s", conn.RemoteAddr().String(), message)

		if message == "" {
			continue // abaikan pesan kosong
		}

		// broadcast message ke semua client selain sender
		formattedMsg := fmt.Sprintf("%s: %s", username, message)
		broadcastMessage(conn, formattedMsg)
	}
}

// fungsi untuk validasi username client
func checkUsername(conn net.Conn, reader *bufio.Reader) (string, error) {
	// kirim request server message ke client
	_, err := conn.Write([]byte("Enter username:\n"))
	if err != nil {
		return "", err
	}

	for {
		// baca input username dari client
		username, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		// trim spasi di akhir
		username = strings.TrimSpace(username)

		// case jika username kosong
		if username == "" {
			conn.Write([]byte("Username cannot be empty.\n"))
		} else {
			// cek semua username yang aktif
			clientsMu.Lock()
			isTaken := false
			for _, existUsername := range clients {
				if existUsername == username {
					isTaken = true // username duplikat
					break
				}
			}
			clientsMu.Unlock()

			// jika username sudah diambil, kirim request ke client, minta username baru
			if isTaken {
				conn.Write([]byte("Username is already taken, try another one. Enter username:\n"))
			} else {
				// username valid, user bisa masuk ke sistem chat
				conn.Write([]byte("Welcome, type \"exit\" to close the program.\n"))
				return username, nil
			}
		}
	}
}

// fungsi untuk broadcast message
func broadcastMessage(sender net.Conn, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// cek untuk semua client
	for client := range clients {
		// jika client bukan sender, send message ke client
		if client != sender {
			_, err := client.Write([]byte(message))

			// error jika koneksi ke client terputus
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}
