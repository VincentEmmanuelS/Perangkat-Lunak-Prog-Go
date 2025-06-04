package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":2000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server started. Waiting for clients...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())
		go handleClient(conn)
		// conn.Close() // langsung tutup koneksi, hanya untuk cek koneksi saja
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Client disconnected: %s\n", conn.RemoteAddr().String())
			return
		}
		fmt.Printf("Received message from %s: %s", conn.RemoteAddr().String(), message)
	}
}
