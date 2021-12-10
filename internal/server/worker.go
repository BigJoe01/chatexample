package server

import (
	"net"
	"time"
)

const (
	// connectionDefaultBufferSize contains the value of allocated buffer size for receive buffer
	connectionDefaultBufferSize = 16384
)

// serverLog are common log messages for go routines
const (
	serverLogDispatchStarted  = "Dispatch started"
	serverLogDispatchStopped  = "Dispatch stopped"
	serverLogReceiverStarted  = "Receiver started"
	serverLogReceiverStopped  = "Receiver stopped"
	serverLogTransformStarted = "Transform started"
	serverLogTransformStopped = "Transform stopped"
	serverLogSenderStarted    = "Sender started"
	serverLogSenderStopped    = "Sender stopped"
)

// intDispatchConnection handle incoming connections put connections into the pool
func intDispatchConnection(s *ChatServer) {
	s.wg.Add(1)
	logger.Printf(serverLogDispatchStarted)

	defer func() {
		s.wg.Done()
		logger.Printf(serverLogDispatchStopped)
	}()

	add := func(conn net.Conn) {
		s.connections.PoolMu.Lock()
		defer s.connections.PoolMu.Unlock()
		s.connections.Pool = append(s.connections.Pool,
			&Connection{
				Name:         conn.RemoteAddr().String(),
				Conn:         conn,
				Disconnected: false,
			})
		logger.Printf("Added connection : %s", conn.RemoteAddr().String())
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			_ = s.server.SetDeadline(time.Now().Add(500 * time.Millisecond))
			nc, err := s.server.Accept()
			if err == nil {
				add(nc)
			}
		}
	}
}

// intReceiver receive message from connection pool and check connection timeouts.
// The incoming message sent to message transformation handler
func intReceiver(s *ChatServer) {
	s.wg.Add(1)
	logger.Printf(serverLogReceiverStarted)

	defer func() {
		s.wg.Done()
		logger.Printf(serverLogReceiverStopped)
	}()

	buff := make([]byte, connectionDefaultBufferSize)
	timeout := false

	remove := func() {
		s.connections.PoolMu.Lock()
		defer s.connections.PoolMu.Unlock()
		var nc []*Connection
		for _, v := range s.connections.Pool {
			if !v.Disconnected {
				nc = append(nc, v)
			} else {
				logger.Printf("Removed connection : %s", v.Conn.LocalAddr().String())
			}
		}
		s.connections.Pool = nc
	}

	receive := func() {
		s.connections.PoolMu.RLock()
		defer s.connections.PoolMu.RUnlock()
		var err error
		var count int
		for _, v := range s.connections.Pool {
			if !v.Disconnected {
				_ = v.Conn.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
				count, err = v.Conn.Read(buff)
				if err != nil {
					if cErr, ok := err.(net.Error); ok && !cErr.Temporary() && cErr.Timeout() {
						v.Disconnected = true
						timeout = true
					}
					continue
				}
				if count > 0 {
					s.receiver <- Message{
						Conn: v.Conn,
						Data: buff[0:count],
					}
				}
			}
		}
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			receive()
			if timeout {
				remove()
				timeout = false
			}
		}
	}
}

// intMessageTransform is handle incoming messages
func intMessageTransform(s *ChatServer) {
	s.wg.Add(1)
	logger.Printf(serverLogTransformStarted)

	defer func() {
		s.wg.Done()
		logger.Printf(serverLogTransformStopped)
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.receiver:
			logger.Printf("incoming message from %s [%s]\n", msg.Conn.RemoteAddr().String(), string(msg.Data))
			s.sender <- msg
		}
	}
}

// intSender broadcast messages to all connected client except the sender
func intSender(s *ChatServer) {
	s.wg.Add(1)
	logger.Printf(serverLogSenderStarted)

	defer func() {
		s.wg.Done()
		logger.Printf(serverLogSenderStopped)
	}()

	broadcast := func(m *Message) {
		s.connections.PoolMu.RLock()
		defer s.connections.PoolMu.RUnlock()
		var err error
		for _, v := range s.connections.Pool {
			if !v.Disconnected && v.Conn != m.Conn {
				_, err = v.Conn.Write(m.Data)
				if err != nil {
					addr := m.Conn.RemoteAddr().String()
					v.Disconnected = true
					logger.Printf("Connection timeout: %s", addr)
				}
			}
		}
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.sender:
			broadcast(&msg)
		}
	}
}
