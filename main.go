package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
)

var clients map[int]*Client
var motd string

// main ...
func main() {
	var useTLS bool
	var port int
	var listener net.Listener

	// command-line arguments
	flag.BoolVar(&useTLS, "tls", false, "Enable Transport Layer Security (TLS)")
	flag.IntVar(&port, "port", 6000, "Sets the desired port number for this server")
	flag.StringVar(&motd, "motd", "etc/motd", "Path to motd flat file")
	flag.Parse()

	// start listening
	if useTLS {
		listener = getTLSListener(port)
	} else {
		listener = getTCPListener(port)
	}
	log.Printf("Listening on port %v\n", port)

	// go channels
	messageChannel := make(chan string)
	incomingClient := make(chan *Client)
	removeClient := make(chan *Client)

	// all clients including connections
	clients = make(map[int]*Client)

	go channelHandler(messageChannel, incomingClient, removeClient)

	for {
		conn, err := listener.Accept()
		if err != nil {
			handleError(err)
			continue
		}
		go handleConnection(conn, messageChannel, incomingClient, removeClient)
	}
}

// getTCPListener ...
func getTCPListener(port int) net.Listener {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal(err)
	}

	return listener
}

// getTLSListener ...
func getTLSListener(port int) net.Listener {
	conf, _ := getTLSConfig()
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%v", port), &conf)
	if err != nil {
		log.Fatal(err)
	}
	return listener
}

// handleConnection ...
func handleConnection(conn net.Conn, messageChannel chan string, incomingClient chan *Client, removeClient chan *Client) {

	clientMsgChannel := make(chan string)
	client := newClient(conn, clientMsgChannel)
	incomingClient <- &client

	serverMsgChannel := make(chan string)
	server := ClientServer{}
	go server.serve(&client, serverMsgChannel, motd)

LOOP:
	for {
		select {
		case msg, ok := <-serverMsgChannel:
			if !ok {
				break LOOP
			}
			messageChannel <- msg
		case msg := <-clientMsgChannel:
			_, err := conn.Write([]byte(msg))
			if err != nil {
				break LOOP
			}
		}
	}

	conn.Close()
	log.Printf("Connection from %v closed.\n", conn.RemoteAddr())
	removeClient <- &client
}
