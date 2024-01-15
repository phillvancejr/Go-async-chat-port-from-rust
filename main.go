package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
	"strings"
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

	// accept clients
	go func() {
		for {
			client, e := server.Accept()
			
			if e != nil {
				// server is closed
				if strings.Contains(e.Error(), "use of closed network connection") {
					return
				} else {
					fmt.Println("error!:", e)
					continue
				}
			}
			client.Write([]byte("Welcome to Go Chat\nenter .exit to leave\n\n"))

			go func() {
				lock.Lock()
				id := client_ids
				client_ids++

				id_string := strconv.Itoa(id)

				clients[id] = Client{make(chan string), client}

				fmt.Printf("client %v connected!\n", id)
				fmt.Println("total clients:", len(clients))

				received_messages := clients[id].channel
				lock.Unlock()

				// buffered reader to read line
				reader := bufio.NewReader(client)

				// channel to receive the message from sender
				input := make(chan string)
				// read input
				go func() {
					for {
						line, _, _ := reader.ReadLine()

						input <- string(line) + "\n"
					}
				}()

				for {

					select {
					case sent := <-input:
						// quit condition
						if strings.HasPrefix(sent, ".exit") {
							quit <- id
							return
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

// broadcast
	for {
		select {
		case id := <-quit:
			println("client", id, "quit!")
			lock.Lock()
			client := clients[id]
			// goodbye
			client.conn.Write([]byte("goodbye!\n"))
			// close connection
			client.conn.Close()
			// delete client
			delete(clients, id)

			if len(clients) == 0 {
				lock.Unlock()
				println("shutdown")
				return
			}
			lock.Unlock()
		case msg := <-messages:
			dark := "\x1b[90m"
			lock.Lock()
			for id, client := range clients {
				// default color/white
				color := "\x1b[0m"
				// send darkened messge back to sender
				if id == msg.id {
					color = dark
				}
				client.channel <- color + msg.msg + "\x1b[0m"
			}
			lock.Unlock()
			// display message on server side
			fmt.Printf(dark + msg.msg + "\x1b[0m")

		}
	}
}
