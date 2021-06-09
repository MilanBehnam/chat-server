package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var clients = make(map[string]net.Conn)
var leaving = make(chan message)
var messages = make(chan message)

type message struct {
	text    string
	address string
}

func main() {
	listen, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handle(conn)
	}

}
func handle(conn net.Conn) {
	clients[conn.RemoteAddr().String()] = conn

	messages <- newMessage(" joined.", conn)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- newMessage(": "+input.Text(), conn)
	}

	delete(clients, conn.RemoteAddr().String())

	leaving <- newMessage(" has left.", conn)

	conn.Close() // NOTE: ignoring network errors
}

func newMessage(msg string, conn net.Conn) message {
	addr := conn.RemoteAddr().String()
	return message{
		text:    addr + msg,
		address: addr,
	}
}

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			for _, conn := range clients {
				if msg.address == conn.RemoteAddr().String() {
					continue
				}
				fmt.Fprintln(conn, msg.text) // NOTE: ignoring network errors
			}

		case msg := <-leaving:
			for _, conn := range clients {
				fmt.Fprintln(conn, msg.text) // NOTE: ignoring network errors
			}

		}
	}
}
