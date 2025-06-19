/*
Dibuat untuk memenuhi tugas besar mata kuliah AIF233131 Pemrograman Go

Kelompok B:
@Author Vincent Emmanuel Suwardy / 6182201067
@Contributor Michael William Iswadi / 6182201019
@Contributor Stanislaus Nathan / 6182201092
@Contributor Jensen Hiem / 6182201026
*/

/* Referensi:
- https://pkg.go.dev/net
- https://go.dev/blog/defer-panic-and-recover
- https://pkg.go.dev/strconv
- https://pkg.go.dev/strings
- https://pkg.go.dev/sync
- https://www.w3schools.com/go/
- https://www.slingacademy.com/article/using-the-net-package-for-low-level-network-programming-in-go/
- Semua slide kuliah
*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// connect ke server
	conn, err := net.Dial("tcp", "127.0.0.1:2000")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	// fmt.Println("Connected to server successfully.")

	// reader untuk baca input username
	reader := bufio.NewReader(os.Stdin)

	// reader untuk baca koneksi server
	connReader := bufio.NewReader(conn)

	// looping untuk cek username availability sampai username accept
	for {
		// read message dari server -> "Enter username: ""
		serverMsg, err := connReader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server")
			return
		}
		fmt.Print(serverMsg)

		// baca username
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Fprintf(conn, username+"\n") // send username ke server
		break

	}

	// Goroutine untuk menerima dan menampilkan pesan dari server
	go func() {
		reader := bufio.NewReader(conn)
		for {
			// read message dari server
			message, err := reader.ReadString('\n')

			if err != nil {
				fmt.Println("Disconnected from server.")
				os.Exit(0) // exit jika disconnected
			}

			// show message dari broadcast server
			// fmt.Print("Message from server: " + message)
			fmt.Print(message)
		}
	}()

	// input dari user untuk dikirim ke server
	for {
		// fmt.Print("Enter message: ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message) // triming spasi

		if message == "" {
			continue // abaikan pesan kosong
		}

		// jika user ketik "exit", keluar dari program -> sementara aja kyknya(?)
		if message == "/exit" {
			fmt.Println("Exiting the program...")
			break
		}

		// handle perintah /join dengan password
		if strings.HasPrefix(message, "/join ") {
			parts := strings.Fields(message)
			if len(parts) < 2 {
				fmt.Println("> Invalid command. Use: /join <room_name> [password]")
				continue
			}
			roomName := parts[1]
			password := ""
			if len(parts) > 2 {
				password = parts[2]
			}
			fmt.Fprintf(conn, "/join %s %s\n", roomName, password)
			continue
		}

		// send message ke server
		fmt.Fprintf(conn, message+"\n")
	}
}
