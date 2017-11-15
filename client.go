package main

import (
	"net"
)

// Client ...
type Client struct {
	ID         int
	Connection net.Conn
	Channel    chan<- string
	Nickname   string
	Away       bool
	AwayMsg    string
}

// newClient ...
func newClient(conn net.Conn, clientChannel chan string) (newClient Client) {
	newClient = Client{0, conn, clientChannel, "nobody", false, ""}
	return newClient
}
