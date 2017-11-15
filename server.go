package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

// ClientServer is the JRC server.
//
type ClientServer struct {
	User           *Client
	MessageChannel chan string
}

// serve is the primary goroutine for this server application.
// A client session begins here.
func (cs *ClientServer) serve(client *Client, serverMsgChannel chan string, motdPath string) {
	cs.User = client
	cs.MessageChannel = serverMsgChannel

	defer close(cs.MessageChannel)

	bufc := bufio.NewReader(cs.User.Connection)
	ident, _, err := bufc.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	if !cs.identIsValid(string(ident)) {
		return
	}

	cs.motd(motdPath)
	usersMessage := fmt.Sprintf("%v user(s) connected.\r\nUSERS: %v\r\n\n", cs.numUsers(), cs.listUsers())
	cs.User.Connection.Write([]byte(usersMessage))
	cs.startSession(bufc)
}

// identIsValid returns a boolean if the proper identification
// command is given.
func (cs *ClientServer) identIsValid(identString string) bool {
	if !strings.HasPrefix(identString, "ident") {
		return false
	}

	identChunk := strings.Split(identString, " ")
	cs.User.Nickname = identChunk[1]

	return true
}

// motd reads the current motd from a flat file.
// motdPath is the path to the file.
func (cs *ClientServer) motd(motdPath string) {
	motd, err := ioutil.ReadFile(motdPath)
	if err != nil {
		handleError(err)
	}
	motd = []byte(strings.Replace(string(motd), "$nickname", cs.User.Nickname, -1))
	cs.User.Connection.Write(motd)
}

// numUsers returns the number of connected clients.
//
func (cs *ClientServer) numUsers() int {
	return len(clients)
}

// listUsers returns a list of all client nicknames.
//
func (cs *ClientServer) listUsers() string {
	var users string

	for _, user := range clients {
		users = fmt.Sprintf("%v %v", users, user.Nickname)
	}
	return strings.TrimSpace(users)
}

// startSession initiates the input loop for the connected clients
// commands.
func (cs *ClientServer) startSession(bufc *bufio.Reader) {
	cs.MessageChannel <- fmt.Sprintf("*** %v has joined the conversation.\r\n", cs.User.Nickname)
	for {
		rawCommand, _, err := bufc.ReadLine()
		if err != nil {
			break
		}
		cs.parseCommand(rawCommand)
	}
	cs.MessageChannel <- fmt.Sprintf("*** %v has left the conversation.\r\n", cs.User.Nickname)
}

// parseCommand ...
func (cs *ClientServer) parseCommand(rawCommand []byte) {
	parms := strings.Split(string(rawCommand), " ")

	switch parms[0] {
	case "MSG":
		cs.handleMessage(strings.TrimPrefix(string(rawCommand), "MSG "))
	case "ACTION":
		cs.handleAction(strings.TrimPrefix(string(rawCommand), "ACTION "))
	case "NICK":
		if len(parms) > 1 {
			cs.changeNick(parms[1])
		}
	case "AWAY":
		if len(parms) > 1 {
			cs.toggleAway(strings.TrimPrefix(string(rawCommand), "AWAY "))
		} else {
			cs.toggleAway("")
		}
	case "URL":
		if len(parms) > 2 {
			cs.handleSendURL(parms[1] /* target */, parms[2] /* url */)
		}
	default:
		cs.handleMessage(string(rawCommand))
	}
}

// handleMessage ...
func (cs *ClientServer) handleMessage(msg string) {
	if cs.User.Away {
		cs.User.Channel <- "- You are away.\r\n"
		return
	}

	cs.MessageChannel <- fmt.Sprintf("%v: %v\r\n", cs.User.Nickname, msg)
}

// handleAction ...
func (cs *ClientServer) handleAction(msg string) {
	cs.MessageChannel <- fmt.Sprintf("* %v %v\r\n", cs.User.Nickname, msg)
}

// changeNick ...
func (cs *ClientServer) changeNick(nickname string) {
	for _, client := range clients {
		if client.Nickname == nickname {
			cs.User.Channel <- fmt.Sprintf("- The nickname %v is already in use!\r\n", nickname)
			return
		}
	}

	cs.MessageChannel <- fmt.Sprintf("*** %v is now known as %v\r\n", cs.User.Nickname, nickname)
	cs.User.Nickname = nickname
}

// toggleAway ...
func (cs *ClientServer) toggleAway(msg string) {
	if cs.User.Away {
		cs.User.Away = false
		cs.MessageChannel <- fmt.Sprintf("** %v is no longer away\r\n", cs.User.Nickname)
		// cs.User.Channel <- "- You are no longer marked as away.\r\n"
	} else {
		cs.User.Away = true
		cs.MessageChannel <- fmt.Sprintf("** %v is now away\r\n", cs.User.Nickname)
		// cs.User.Channel <- "- You are now marked as away.\r\n"
	}
	cs.User.AwayMsg = msg
}

// handleSendURL ...
func (cs *ClientServer) handleSendURL(target string, url string) {
	for _, t := range clients {
		if t.Nickname == target {
			if t.Away {
				cs.User.Channel <- fmt.Sprintf("%v is away.\r\n", target)
				return
			}
			t.Channel <- fmt.Sprintf("open.url|%v|%v\r\n", cs.User.Nickname, url)
			cs.User.Channel <- fmt.Sprintf("- Sent url to %v\r\n", target)
			return
		}
	}
	cs.User.Channel <- fmt.Sprintf("- %v does not exist.\r\n", target)
}
