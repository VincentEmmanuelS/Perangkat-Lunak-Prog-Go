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

		// send message ke server
		fmt.Fprintf(conn, message+"\n")
	}
}
