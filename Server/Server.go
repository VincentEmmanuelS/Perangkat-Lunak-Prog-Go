package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"strconv"
)

// struct client buat chat room
type Client struct {
	conn net.Conn // koneksi TCP client
	name string   // username
	room *Room    // pointer ke room
}

// Chat room struct
type Room struct {
	name    string           // room name
	size 	int 	  			 // max size
	clients map[*Client]bool // list client di dalam room
	mu      sync.Mutex       // mutex untuk map Room.clients
}

// Map untuk menyimpan koleksi client yg aktif
var (
	clientsMu sync.Mutex // mutex untuk map clients
	// clients   = make(map[net.Conn]string)
	clients = make(map[*Client]bool)

	roomsMu sync.Mutex // mutex untuk map rooms
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

		// new client dari connection
		client := &Client{conn: conn}
		clientsMu.Lock()
		clients[client] = true
		clientsMu.Unlock()

		// handle client pada goroutine terpisah
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
	// send notif ke client, ketika ada client baru yang connect
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
		// broadcastMessage(nil, fmt.Sprintf("%s has left!\n", username))
	}()

	// show available room
	client.conn.Write([]byte("\n"))
	client.conn.Write([]byte("====================\n"))
	client.conn.Write([]byte("Available room:\n"))
	listRooms(client)

	client.conn.Write([]byte("====================\n"))
	client.conn.Write([]byte("Commands: \n"))
	client.conn.Write([]byte("/create <room_name> [room_size]\n"))
	client.conn.Write([]byte("/join <room_name>\n"))
	client.conn.Write([]byte("/roomlist\n"))
	client.conn.Write([]byte("/exit\n"))
	client.conn.Write([]byte("====================\n"))
	client.conn.Write([]byte("\n"))

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
		if strings.HasPrefix(message, "/create ") {
			// create room baru
			roomName := strings.TrimSpace(strings.TrimPrefix(message, "/create "))
			words := strings.Split(roomName, " ")
			size := -1
			roomName = words[0]

			// Validasi -> tidak boleh ada spasi untuk nama room
			if len(words) > 1  {
				if num, err := strconv.Atoi(words[1]); err != nil || len(words) > 2{
					client.conn.Write([]byte("> Room name cannot contain spaces. Use underscores (_) or a single word.\n"))
					client.conn.Write([]byte("\n"))
					continue
				} else {
					size = num
				}
			}

			client.createRoom(roomName, size)
			broadcastMessage(client, "> New room '"+roomName+"' is created!\n")
			client.joinRoom(roomName)
			// fmt.Printf("%s create %s\n", client.name, client.room.name)
			continue
		} else if strings.HasPrefix(message, "/join ") {
			// join available room
			roomName := strings.TrimSpace(strings.TrimPrefix(message, "/join "))
			client.joinRoom(roomName)
			// fmt.Printf("%s join %s\n", client.name, client.room.name)
			continue
		} else if message == "/roomlist" {
			// show available room
			client.conn.Write([]byte("====================\n"))
			client.conn.Write([]byte("Available rooms:\n"))
			listRooms(client)
			client.conn.Write([]byte("====================\n"))
			client.conn.Write([]byte("\n"))
			continue
		} else if message == "/leave" {
			if client.room != nil {
				client.room.removeClient(client)
				client.conn.Write([]byte("> You have left the room.\n"))
				client.room = nil

				// print format ke client
				client.conn.Write([]byte("\n"))
				client.conn.Write([]byte("====================\n"))
				client.conn.Write([]byte("Available room:\n"))
				listRooms(client)

				client.conn.Write([]byte("====================\n"))
				client.conn.Write([]byte("Commands: \n"))
				client.conn.Write([]byte("/create <room_name>\n"))
				client.conn.Write([]byte("/join <room_name>\n"))
				client.conn.Write([]byte("/roomlist\n"))
				client.conn.Write([]byte("/exit\n"))
				client.conn.Write([]byte("====================\n"))
				client.conn.Write([]byte("\n"))
			} else {
				// handling
				client.conn.Write([]byte("> You are not in a room.\n"))
				client.conn.Write([]byte("\n"))
			}
			continue
		} else if client.room == nil {
			client.conn.Write([]byte("> You must join a room to chat. Type /roomlist to see available room(s).\n"))
			client.conn.Write([]byte("\n"))
		}

		// broadcast message ke room -> kondisi client sudah di dalam room
		if client.room != nil {
			formattedMsg := fmt.Sprintf("%s: %s\n", client.name, message)
			client.room.broadcast(client, formattedMsg)
		}
	}
}

// fungsi untuk broadcast message secara global (client yang belum masuk ke room)
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
			conn.Write([]byte("> Username cannot be empty.\n"))
			conn.Write([]byte("> Enter username:\n"))
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
				conn.Write([]byte("> Username is already taken, try another one.\n"))
				conn.Write([]byte("> Enter username:\n"))
			} else {
				// username valid, user bisa masuk ke sistem chat
				// conn.Write([]byte("Welcome, type \"exit\" to close the program.\n"))
				return username, nil
			}
		}
	}
}

// list semua room yang tersedia
func listRooms(client *Client) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	// jika length = 0 -> maka belum ada room yang dibuat
	if len(rooms) == 0 {
		client.conn.Write([]byte("- No rooms available\n"))
		return
	}

	// print semua room name
	for roomName, room := range rooms {
		if room.size != -1{
			client.conn.Write([]byte("- " + roomName + " " + strconv.Itoa(len(room.clients)) + "/" + strconv.Itoa(room.size)))
		}else {
			client.conn.Write([]byte("- " + roomName))
		}
		// jika client ada di room tersebut
		if (client.room == room){
			client.conn.Write([]byte(" <-"))
		}
		client.conn.Write([]byte("\n"))
	}
}

// create room
func (c *Client) createRoom(name string, size int) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	// cek apakah room sudah tersedia atau belum
	// cek dengan menggunakan nama room -> dalam kasus ini nama room harus unik (tidak ada room dengan nama yang sama lebih dari 1)
	if _, exists := rooms[name]; !exists {
		// create room
		room := &Room{
			name:    name,
			size: 	 size,
			clients: make(map[*Client]bool, size),
		}
		rooms[name] = room
		// fmt.Println(rooms[name])
		c.conn.Write([]byte("> Room " + name + " created!\n"))
	} else {
		c.conn.Write([]byte("> Room " + name + " is already exists\n"))
	}

}

// join room by name
func (c *Client) joinRoom(name string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	// cek apakah room ada atau tidak
	if room, exists := rooms[name]; exists {
		// cek apakah room sudah penuh
		if len(room.clients) >= room.size && room.size != -1{
			c.conn.Write([]byte("> Room is full\n")) 
		} else {
			// jika tujuan client adalah room yand dia sedang berada
			if (c.room == room){
				c.conn.Write([]byte("> You're already in this room\n"))
			} else {
				// jika client sudah berada di 1 room, dan '/join' maka user akan dipindah room ke room yang baru
				if c.room != nil {
					c.room.removeClientNoLock(c)
				}

				room.addClient(c) // join room
				c.room = room
				c.conn.Write([]byte("> You have joined " + name + "\n"))
				c.conn.Write([]byte("> Type '/leave' to leave the chatroom\n"))
				}
		}	
	} else {
		c.conn.Write([]byte("> Room does not exist\n"))
	}
}

// menambahkan client ke room
func (r *Room) addClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client] = true
	// fmt.Printf("%s join %s\n", client.name, r.name)
	r.broadcast(client, fmt.Sprintf("> %s has joined the room.\n", client.name))
}

// mengeluarkan client dari room
func (r *Room) removeClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// fmt.Printf("%s left %s\n", client.name, client.room.name)
	delete(r.clients, client)
	r.broadcast(client, fmt.Sprintf("> %s has left the room.\n", client.name))

	// jika room kosong, hapus dari map global `rooms`
	if len(r.clients) == 0 {
		roomsMu.Lock()
		delete(rooms, r.name)
		delete(rooms, r.name)
		roomsMu.Unlock()

		broadcastMessage(nil, "> Room '"+r.name+"' is deleted, because no one is in this room.\n")
	}
}

func (r *Room) removeClientNoLock(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// fmt.Printf("%s left %s\n", client.name, client.room.name)
	delete(r.clients, client)
	r.broadcast(client, fmt.Sprintf("> %s has left the room.\n", client.name))

	// jika room kosong, hapus dari map global `rooms`
	if len(r.clients) == 0 {
		// no lock here
		delete(rooms, r.name)
		delete(rooms, r.name)

		broadcastMessage(nil, "> Room '"+r.name+"' is deleted, because no one is in this room.\n")
	}
}

// broadcast message di dalam room kecuali ke sender
func (r *Room) broadcast(sender *Client, message string) {
	for client := range r.clients {
		if client != sender {
			_, err := client.conn.Write([]byte(message))

			// error jika koneksi ke client terputus
			if err != nil {
				client.conn.Close()
				delete(r.clients, client)
			}
		}
	}
}

// ---------------------------------------------------------------------------

// log: 09/06/2025 (author nathan)

// change note: 
// line 22: add attribute "size int" to struct room
// line 134: validation condition changed for room size parameter
// modify listroom to print current num of client in the room / cap
// modify createroom to take in size as parameter

// to do: 
// add default mode to not input size for create room so it's infinite
// modify UI for better description (menu and listroom)
// testing + debug for degenerate cases

// -----------------------------------------------------------------------------

// log: 10/06/2025 (author Nathan)

// - deadlock found in /join if you're in a room, made a seperate remove clien func with no roomMu.lock
// - edit UI: add an arrow to pointout which room you're located in roomlist, change command syntax for create room (optional room cap)
// - handle rejoining a room you're already inside in
