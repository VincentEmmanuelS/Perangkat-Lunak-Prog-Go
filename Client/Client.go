package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:2000")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// fmt.Println("Connected to server successfully.")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message: ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if message == "exit" {
			fmt.Println("Exiting the program...")
			break
		}
		fmt.Fprintf(conn, message+"\n")
	}
}
