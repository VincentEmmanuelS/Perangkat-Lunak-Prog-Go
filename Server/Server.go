package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

// Map untuk menyimpan koleksi client yg aktif
var (
	clients   = make(map[net.Conn]bool)
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

		// add client ke daftar clients aktif
		// di protect Mutex
		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()

		fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())
		go handleClient(conn) // handle comms client pada goroutine terpisah
		// conn.Close() // langsung tutup koneksi, hanya untuk cek koneksi saja
	}
}

// Handle setiap client
func handleClient(conn net.Conn) {
	// defer anonymous class -> dieksekusi ketika client disconnect
	defer func() {
		clientsMu.Lock()
		delete(clients, conn) // hapus client dari daftar aktif
		clientsMu.Unlock()
		conn.Close() // tutup koneksi
		fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
	}()

	// defer conn.Close()
	reader := bufio.NewReader(conn)

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

		// broadcast message ke semua client selain sender
		broadcastMessage(conn, message)
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
