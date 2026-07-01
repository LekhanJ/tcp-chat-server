package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
	send chan string
}

type Room struct {
	mu       sync.Mutex
	members  []*Client
	roomId   string
	password string
}


type Message struct {
	Sender *Client
	Text   string
}

func (r *Room) AddMember(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members = append(r.members, c)
}

func (r *Room) RemoveMember(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, m := range r.members {
		if m == c {
			r.members = append(r.members[:i], r.members[i+1:]...)
			break
		}
	}
}

func (r *Room) Broadcast(msg Message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.members {
		if m == msg.Sender {
			continue
		}
		select {
		case m.send <- msg.Sender.name + ": " + msg.Text:
		default:
			
		}
	}
}


func main() {
	listener, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatal("Could not listen: ", err)
	}
	defer listener.Close()

	room := &Room{
		roomId:   "roomid",
		password: "roompass",
	}

	messages := make(chan Message)

	go func() {
		for msg := range messages {
			room.Broadcast(msg)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		client := &Client{
			conn: conn,
			send: make(chan string, 16),
			name: "Unknown",
		}

		go handleClient(client, room, messages)
	}
}

func handleClient(client *Client, room *Room, messages chan<- Message) {
	defer client.conn.Close()

	reader := bufio.NewReader(client.conn)

	if !handleRoomJoin(client, room, reader) {
		return
	}

	room.AddMember(client)
	defer room.RemoveMember(client)

	go clientWriter(client)

	clientReader(client, reader, messages)
}

func handleRoomJoin(client *Client, room *Room, reader *bufio.Reader) bool {
	if _, err := client.conn.Write([]byte("To join a room use \"JOIN <room-id> <room-password> <your-name>\"\n")); err != nil {
		fmt.Println("Could not send message: ", err)
		return false
	}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Could not read message: ", err)
			return false
		}

		tokens := strings.Fields(message)
		if len(tokens) < 3 || len(tokens) > 4 || tokens[0] != "JOIN" || tokens[1] != room.roomId || tokens[2] != room.password {
			if _, err := client.conn.Write([]byte("Invalid Operation!\n")); err != nil {
				fmt.Println("Could not send message: ", err)
				return false
			}
			continue
		}
		if len(tokens) == 4 {
			client.name = tokens[3]
		}
		break
	}

	if _, err := client.conn.Write([]byte("You Joined the Room\n")); err != nil {
		fmt.Println("Could not send message: ", err)
		return false
	}
	return true
}

func clientReader(client *Client, reader *bufio.Reader, messages chan<- Message) {
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected: ", err)
			return
		}
		messages <- Message{Sender: client, Text: text}
	}
}

func clientWriter(client *Client) {
	for message := range client.send {
		if _, err := client.conn.Write([]byte(message)); err != nil {
			fmt.Println("Could not send message: ", err)
			return
		}
	}
}