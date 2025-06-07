package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// struct client buat chat room
type Client struct {
	conn net.Conn
	name string
	room *Room
}

// Chat room struct
type Room struct {
	name    string
	clients map[*Client]bool
	mu      sync.Mutex
}

// Map untuk menyimpan koleksi client yg aktif
var (
	clientsMu sync.Mutex
	// clients   = make(map[net.Conn]string)
	clients = make(map[*Client]bool)

	roomsMu sync.Mutex
	rooms   = make(map[string]*Room)
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

		// fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())

		client := &Client{conn: conn}
		clientsMu.Lock()
		clients[client] = true
		clientsMu.Unlock()

		// handle comms client pada goroutine terpisah
		// go handleClient(conn)
		go handleClient(client)
	}
}

func handleClient(client *Client) {
	reader := bufio.NewReader(client.conn)

	// read username dari client
	username, err := checkUsername(client.conn, reader)
	if err != nil {
		client.conn.Close()
		return
	}
	// username = strings.TrimSpace(username)

	// simpan username dari client
	client.name = username
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	fmt.Printf("%s connected\n", username)
	// // send notif ke client, ketika ada client baru yang connect
	// formattedNotif := fmt.Sprintf("%s jumped in!\n", username)
	// broadcastMessage(client, formattedNotif)

	defer func() {
		client.conn.Close()
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()

		if client.room != nil {
			client.room.removeClient(client)
		}
		fmt.Printf("%s disconnected\n", username)
		broadcastMessage(nil, fmt.Sprintf("%s has left!\n", username))
	}()

	// show available room
	client.conn.Write([]byte("Available room:\n"))
	listRooms(client)

	for {
		// read message dari client
		message, err := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if err != nil {
			// Client disconnected or error
			break
		}
		// fmt.Printf("Received message from %s: %s", conn.RemoteAddr().String(), message)

		if message == "" {
			continue // abaikan pesan kosong
		}

		// chatroom commands
		if client.room == nil {
			if strings.HasPrefix(message, "/create ") {
				roomName := strings.TrimSpace(strings.TrimPrefix(message, "/create "))
				// fmt.Println(roomName)
				client.createRoom(roomName)
				// fmt.Printf("%s create %s\n", client.name, client.room.name)
				continue
			} else if strings.HasPrefix(message, "/join ") {
				roomName := strings.TrimSpace(strings.TrimPrefix(message, "/create "))
				client.joinRoom(roomName)
				// fmt.Printf("%s join %s\n", client.name, client.room.name)
				continue
			} else if message == "/rooms" {
				client.conn.Write([]byte("Available rooms:\n"))
				listRooms(client)
				continue
			} else if message == "/leave" {
				client.conn.Write([]byte("You are not in a room.\n"))
			} else {
				client.conn.Write([]byte("You must join a room to chat. Type /rooms to see availavble room(s).\n"))
			}
		}

		if message == "/leave" {
			if client.room != nil {
				client.room.removeClient(client)
				client.conn.Write([]byte("You have left the room.\n"))
				client.room = nil
			} else {
				client.conn.Write([]byte("You are not in a room.\n"))
			}
			continue
		}

		if client.room != nil {
			formattedMsg := fmt.Sprintf("%s: %s\n", client.name, message)
			client.room.broadcast(client, formattedMsg)
		}
	}
}

/*
// Handle setiap client
func old_handleClient(conn net.Conn) {
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
*/

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
			conn.Write([]byte("Username cannot be empty. Enter username:\n"))
		} else {
			// cek semua username yang aktif
			clientsMu.Lock()
			isTaken := false
			for client := range clients {
				if client.name == username {
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
				// conn.Write([]byte("Welcome, type \"exit\" to close the program.\n"))
				return username, nil
			}
		}
	}
}

/*
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
*/

// fungsi untuk broadcast message
func broadcastMessage(sender *Client, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// cek untuk semua client
	for client := range clients {
		// jika client bukan sender, send message ke client
		if client != sender {
			_, err := client.conn.Write([]byte(message))

			// error jika koneksi ke client terputus
			if err != nil {
				client.conn.Close()
				delete(clients, client)
				// remove client dari room
				if client.room != nil {
					client.room.removeClient(client)
				}
			}
		}
	}
}

// list semua room yang tersedia
func listRooms(client *Client) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if len(rooms) == 0 {
		client.conn.Write([]byte("- No rooms available\n"))
		return
	}

	for roomName := range rooms {
		client.conn.Write([]byte("- " + roomName + "\n"))
	}
}

// create room
func (c *Client) createRoom(name string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if _, exists := rooms[name]; !exists {
		room := &Room{
			name:    name,
			clients: make(map[*Client]bool),
		}
		rooms[name] = room
		fmt.Println(rooms[name])
		c.conn.Write([]byte("Room created: " + name + "\n"))
	} else {
		c.conn.Write([]byte("Room already exists: " + name + "\n"))
	}
}

// join room by name
func (c *Client) joinRoom(name string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	// room, exists := rooms[name]
	if room, exists := rooms[name]; exists {
		if c.room != nil {
			c.room.removeClient(c) // Leave current room if any
		}
		room.addClient(c) // Join the new room
		c.room = room
		c.conn.Write([]byte("Joined room: " + name + "\n"))
	} else {
		c.conn.Write([]byte("Room does not exist\n"))
	}

	// if exists {
	// 	if c.room != nil {
	// 		c.room.removeClient(c)
	// 	}

	// 	rooms[name].addClient(c)
	// 	c.room = rooms[name]
	// 	c.conn.Write([]byte("Joined room: " + name + "\n"))
	// } else {
	// 	c.conn.Write([]byte("Room does not exist\n"))
	// }
	// if !exists {
	// 	room = &Room{
	// 		name:    name,
	// 		clients: make(map[*Client]bool),
	// 	}
	// 	rooms[name] = room
	// }

}

// menambahkan client ke room
func (r *Room) addClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client] = true
	fmt.Printf("%s join %s\n", client.name, r.name)
	r.broadcast(nil, fmt.Sprintf("%s has joined the room.\n", client.name))
}

// mengeluarkan client dari room
func (r *Room) removeClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Printf("%s left %s\n", client.name, client.room.name)
	delete(r.clients, client)
	r.broadcast(nil, fmt.Sprintf("%s has left the room \n", client.name))
}

// broadcast message di dalam room kecuali ke sender
func (r *Room) broadcast(sender *Client, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for client := range r.clients {
		if client != sender {
			_, err := client.conn.Write([]byte(message))

			if err != nil {
				client.conn.Close()
				delete(r.clients, client)
			}
		}
	}
}
