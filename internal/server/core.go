// Server implements simple tcp/ip server what can handle incoming chat data.
// I prefer structure based composition instead of use full static variable and methods
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type (

	// Message represents internal message between receiver, message transform and sender
	Message struct {
		Conn net.Conn
		Data []byte
	}

	// Connection represents one user connection and user details
	Connection struct {
		Name         string
		Conn         net.Conn
		Disconnected bool
	}

	// Connections represents list of the connected chat clients
	Connections struct {
		PoolMu sync.RWMutex
		Pool   []*Connection
	}

	// Chat server represents the running tcp/ip server details
	ChatServer struct {
		port        int
		running     bool
		wg          sync.WaitGroup
		connections Connections
		server      *net.TCPListener
		ctx         context.Context
		ctxCancel   func()
		receiver    chan Message
		sender      chan Message
	}
)

var (
	logger *log.Logger
)

const (
	// serverSettingsAddress contains the default local ip address
	serverSettingsAddress = "0.0.0.0"
)

// serverLog messages used for logging server states
const (
	serverLogStarting = "Server starting"
	serverLogStarted  = "Server started"
	serverLogStopping = "Server stopping"
	serverLogStopped  = "Server stopped"
)

// Error codes returned by start, stop functions
var (
	ErrServerAlreadyRunning = errors.New("Server already running")
	ErrNotRunning           = errors.New("Couldn't stop the server, server not running")
)

// NewChatBroadcastServer creates new chat server
func NewChatBroadcastServer(p int, l *log.Logger) *ChatServer {
	s := &ChatServer{
		port:     p,
		receiver: make(chan Message, 100),
		sender:   make(chan Message, 100),
		running:  false,
	}
	logger = l
	return s
}

// Start is allocate specified port for the server and try to run the background message handler
func (s *ChatServer) Start() error {
	if s.running {
		return ErrServerAlreadyRunning
	}
	logger.Println(serverLogStarting)
	srv, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serverSettingsAddress, s.port))
	if err != nil {
		return err
	}
	s.server, _ = srv.(*net.TCPListener)
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	go intMessageTransform(s.ctx, &s.wg, s.receiver, s.sender)
	go intAccept(s)
	go intSender(s)
	go intReceiver(s)

	s.running = true
	logger.Println(serverLogStarted)
	return nil
}

// Stop is sending stop signal for the running server
func (s *ChatServer) Stop() error {
	if !s.running {
		return ErrNotRunning
	}
	logger.Println(serverLogStopping)
	s.ctxCancel()
	s.wg.Wait()

	for _, v := range s.connections.Pool {
		_ = v.Conn.Close()
	}
	err := s.server.Close()
	close(s.receiver)
	close(s.sender)
	logger.Println(serverLogStopped)
	return err
}

// IsRunning it returns the server current status
func (s *ChatServer) IsRunning() bool {
	return s.running
}

// Port returns server default port
func (s *ChatServer) Port() int {
	return s.port
}
