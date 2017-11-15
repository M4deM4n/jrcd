package main

func channelHandler(
	messageChannel <-chan string,
	incomingClient <-chan *Client,
	removeClient <-chan *Client,
) {
	for {
		select {
		case msg := <-messageChannel:
			for _, c := range clients {
				go func(mch chan<- string) { mch <- msg }(c.Channel)
			}

		case client := <-incomingClient:
			id := len(clients)
			client.ID = id
			clients[id] = client

		case client := <-removeClient:
			delete(clients, client.ID)
		}
	}
}
