// Chat client application example
// This example it creates few running chat client, this clients send random message to the server
package main

import (
	"chat/internal/client"
	"log"
	"os"
	"time"
)

const (
	// serverAddress contains the chat server address
	serverAddress = "localhost:8080"
	// clientCount represents how many client created for the client test
	clientCount = 3
)

var (
	// logger is the default std out logger with date
	logger  *log.Logger = log.New(os.Stdout, "", log.LstdFlags)
	clients []*client.ChatClient
)

func main() {
	client.SetLogger(logger)
	var cl *client.ChatClient
	var err error

	for c := 0; c < clientCount; c++ {
		cl, err = client.NewClient(serverAddress, time.Second*2)
		if err == nil {
			clients = append(clients, cl)
			_ = cl.Start()
		}
	}

	ticker := time.NewTicker(time.Second * 5)
	<-ticker.C

	for _, v := range clients {
		_ = v.Stop()
	}

}
