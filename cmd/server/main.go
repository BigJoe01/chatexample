// Chat server application
package main

import (
	"chat/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	// basePort contains the running server default port
	basePort = 8080
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	s := server.NewChatBroadcastServer(basePort, logger)
	err := s.Start()
	if err != nil {
		logger.Println(err)
		return
		// os.Exit(1) sometime the developers don't like this
	}

	osStopChannel := make(chan os.Signal)
	signal.Notify(osStopChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-osStopChannel

	err = s.Stop()
	if err != nil {
		logger.Println(err)
	}
}
