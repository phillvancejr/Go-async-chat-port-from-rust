package main

import (
	"bufio"
	"net"
	"strconv"
	"sync"
)

type Message struct {
	id  int
	msg string
}

type Client struct {
	channel chan string
	conn    net.Conn
}

var (
	lock       sync.Mutex
	client_ids = 0
	clients    = make(map[int]Client)
)

func main() {
	server, _ := net.Listen("tcp", ":8000")
	defer server.Close()

	// broadcast channel
	messages := make(chan Message)
	// quit channel
	quit := make(chan int)

	// setup sockets
	go func() {
		for {
			client, e := server.Accept()
			if e != nil {
				continue
			}
			client.Write([]byte("Welcome to Chat\n"))

			go func() {
				lock.Lock()
				id := client_ids
				client_ids++
				lock.Unlock()

				id_string := strconv.Itoa(id)

				clients[id] = Client{make(chan string), client}

				println("client", id, "connected!")
				println("total clients:", len(clients), "\n")

				received_messages := clients[id].channel

				// buffered reader to read line
				reader := bufio.NewReader(client)

				// channel to receive the message from sender
				input := make(chan string)

			chat:
				for {
					// read input
					go func() {
						line, _, _ := reader.ReadLine()

						input <- string(line) + "\n"
					}()

					select {
					case sent := <-input:
						//println("sent:", sent, "len:", len(sent))
						if len(sent) == 1 {
							quit <- id
							break chat
						}

						// send message
						messages <- Message{id, id_string + ": " + sent}
					case received := <-received_messages:
						// sent to client
						client.Write([]byte(received))
					}
				}

			}()
		}

	}()

	println("Chat on port 8000")
broadcast:
	for {
		select {
		case id := <-quit:
			println("client", id, "quit!")
			client := clients[id]
			// goodbye
			client.conn.Write([]byte("goodbye!"))
			// close connection
			client.conn.Close()
			// delete client
			delete(clients, id)

			if len(clients) == 0 {
				break broadcast
			}
		case msg := <-messages:
			for id, client := range clients {
				// default color/white
				color := "\x1b[0m"
				// send darkened messge back to sender
				if id == msg.id {
					color = "\x1b[90m"
				}
				client.channel <- color + msg.msg + "\x1b[0m"
			}

		}
	}

	println("shutdown")

}
