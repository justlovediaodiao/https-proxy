package proxy

import (
	"io"
	"net"
	"time"
)

// Conn is a tcp stream connection.
type Conn interface {
	net.Conn
	// Handshake do handshake and return target address of proxied tcp stream.
	Handshake() (string, error)
}

// clientConn is sever side incoming connection.
type clientConn struct {
	net.Conn
	password string
}

// clientConn is client side outgoing connection.
type serverConn struct {
	net.Conn
	targetAddr string
	password   string
}

// socksConn is client side incoming socks connection.
type socksConn struct {
	net.Conn
}

// ClientConn wrap a client side outgoing connection.
func ClientConn(c net.Conn, password string) Conn {
	return &clientConn{c, password}
}

// ServerConn wrap a sever side incoming connection.
func ServerConn(c net.Conn, targetAddr string, password string) Conn {
	return &serverConn{c, targetAddr, password}
}

// SocksConn wrap a client side incoming socks connection.
func SocksConn(c net.Conn) Conn {
	return &socksConn{c}
}

// Handshake do handshake to client.
func (c *clientConn) Handshake() (string, error) {
	return c.handshake()
}

// Handshake do handshake to server.
func (c *serverConn) Handshake() (string, error) {
	return c.targetAddr, c.handshake()
}

// Handshake do handshake to app.
func (c *socksConn) Handshake() (string, error) {
	return c.handshake()
}

// Relay copies between left and right bidirectionally.
func Relay(left, right net.Conn) error {
	ch := make(chan error, 2)

	go func() {
		_, err := io.Copy(right, left)
		ch <- err
		right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	}()

	_, err := io.Copy(left, right)
	ch <- err
	right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	// the first err is relay error reason.
	err = <-ch
	<-ch
	return err
}
