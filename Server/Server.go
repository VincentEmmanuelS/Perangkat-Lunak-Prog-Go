package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Room struct {
	name    string
	clients map[net.Conn]string
}

var (
	rooms   = make(map[string]*Room)
	roomsMu sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":2000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on port 2000...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	conn.Write([]byte("Enter your username:\n"))
	username, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	username = strings.TrimSpace(username)

	var currentRoom *Room
	conn.Write([]byte("Type /create <room>, /join <room>, or /exit:\n"))

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		input = strings.TrimSpace(input)

		if strings.HasPrefix(input, "/create ") {
			roomName := strings.TrimSpace(strings.TrimPrefix(input, "/create "))
			currentRoom = createRoom(roomName)
			if currentRoom != nil {
				joinRoom(currentRoom, conn, username)
				break
			} else {
				conn.Write([]byte("Room already exists. Try another name.\n"))
			}

		} else if strings.HasPrefix(input, "/join ") {
			roomName := strings.TrimSpace(strings.TrimPrefix(input, "/join "))
			currentRoom = joinExistingRoom(roomName, conn, username)
			if currentRoom != nil {
				break
			}

		} else if input == "/exit" {
			return

		} else if input == "/rooms" {
			listRooms(conn)
		} else {
			conn.Write([]byte("Unknown command. Use /create <room>, /join <room>, /rooms, or /exit\n"))
		}
	}

	// Message loop
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			leaveRoom(currentRoom, conn, username)
			return
		}
		msg = strings.TrimSpace(msg)

		if msg == "/leave" {
			leaveRoom(currentRoom, conn, username)
			conn.Write([]byte("You left the room. Use /create or /join again.\n"))
			handleClient(conn)
			return

		} else if msg == "/exit" {
			leaveRoom(currentRoom, conn, username)
			return

		} else if msg != "" {
			broadcastToRoom(currentRoom, conn, fmt.Sprintf("%s: %s\n", username, msg))
		}
	}
}

func createRoom(name string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if _, exists := rooms[name]; exists {
		return nil
	}
	room := &Room{name: name, clients: make(map[net.Conn]string)}
	rooms[name] = room
	return room
}

func joinRoom(room *Room, conn net.Conn, username string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room.clients[conn] = username
	broadcastToRoom(room, conn, fmt.Sprintf("%s joined the room.\n", username))
	conn.Write([]byte(fmt.Sprintf("Room '%s' joined.\n", room.name)))
}

func joinExistingRoom(name string, conn net.Conn, username string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[name]
	if !exists {
		conn.Write([]byte("Room not found. Use /create to make a new one.\n"))
		return nil
	}

	room.clients[conn] = username
	broadcastToRoom(room, conn, fmt.Sprintf("%s joined the room.\n", username))
	conn.Write([]byte(fmt.Sprintf("Room '%s' joined.\n", room.name)))
	return room
}

func leaveRoom(room *Room, conn net.Conn, username string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if room != nil {
		delete(room.clients, conn)
		broadcastToRoom(room, conn, fmt.Sprintf("%s left the room.\n", username))
		if len(room.clients) == 0 {
			delete(rooms, room.name)
		}
	}
}

func broadcastToRoom(room *Room, sender net.Conn, message string) {
	for client := range room.clients {
		if client != sender {
			client.Write([]byte(message))
		}
	}
}

func listRooms(conn net.Conn) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if len(rooms) == 0 {
		conn.Write([]byte("No rooms available.\n"))
		return
	}

	var roomList []string
	for name := range rooms {
		roomList = append(roomList, name)
	}
	conn.Write([]byte("Available rooms: " + strings.Join(roomList, ", ") + "\n"))
}
