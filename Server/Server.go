package main

import (
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
		conn.Close() // langsung tutup koneksi, hanya untuk cek koneksi saja
	}
}
