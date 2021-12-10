// Chat client implementation
package client

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// ChatClient implements tcp client
	ChatClient struct {
		address   string
		connError atomic.Value
		conn      *net.TCPConn
		ctxCancel func()
		ctx       context.Context
		wg        sync.WaitGroup
		running   bool
		interval  time.Duration
	}
)

var (
	// logger is the default logger for the client
	logger *log.Logger
)

// SetLogger is set module wide default logger
func SetLogger(l *log.Logger) {
	logger = l
}

// NewClient is make new chat client
func NewClient(a string, i time.Duration) (*ChatClient, error) {
	c := &ChatClient{
		address:   a,
		running:   false,
		interval:  i,
	}
	c.connError.Store(nil)

	conn, err := net.Dial("tcp", a)
	if err != nil {
		return nil, err
	}
	c.conn = conn.(*net.TCPConn)
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	return c, nil
}

// clientReadWriter is handle client incoming - outgoing messages
func clientReadWriter(client *ChatClient) {
	client.wg.Add(1)
	defer client.wg.Done()

	clientReader := bufio.NewReader(client.conn)
	ticker := time.NewTicker(client.interval)
	for {
		select {
		case <-client.ctx.Done():
			return
		case <-ticker.C:
			_, err := client.conn.Write([]byte(GenerateRandomString(10)))
			switch err {
			case nil:
			case io.EOF:
				logger.Printf("Server closed the connection for %s", client.conn.LocalAddr().String())
				client.connError.Store(err)
				return
			default:
			}
		default:
			_ = client.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			serverData, err := clientReader.ReadString('\n')
			switch err {
			case nil:
				logger.Printf("Server message %s to %s", serverData, client.conn.LocalAddr().String())
			case io.EOF:
				logger.Printf("Server closed the connection for %s", client.conn.LocalAddr().String())
				client.connError.Store(err)
				return
			default:
			}
		}
	}
}

// ErrClient is error messages for client errors
var (
	ErrClientAlreadyRunning = errors.New("Client already running")
)

// Start is create background task what processing incoming and outgoing random messages
func (c *ChatClient) Start() error {
	if c.running {
		return ErrClientAlreadyRunning
	}
	go clientReadWriter(c)
	c.running = true
	return nil
}

// Close is wait for background go routine for shutdown and close the connection
func (c *ChatClient) Stop() error {
	c.ctxCancel()
	c.wg.Wait()
	return c.conn.Close()
}

// IsRunning returns client current state
func (c *ChatClient) IsRunning() bool {
	return c.running
}

// LastError returns connection last error
func (c *ChatClient) LastError() error {
	return c.connError.Load().(error)
}
