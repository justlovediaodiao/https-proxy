package proxy

import (
	"bufio"
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
	key []byte
}

// clientConn is client side outgoing connection.
type serverConn struct {
	net.Conn
	targetAddr string
	key        []byte
}

// socksConn is client side incoming socks connection.
type socksConn struct {
	net.Conn
}

// httpConn is client side incoming http connection.
type httpConn struct {
	net.Conn
	isTunnel bool          // works on tunnel or proxy mode?
	bufConn  *bufio.Reader // used to read net.Conn for http
	request  io.Reader     // http request used to forward to remote
}

// NewClientConn wrap a sever side incoming connection.
func NewClientConn(c net.Conn, password string) Conn {
	return &clientConn{c, kdf(password, 32)}
}

// NewServerConn wrap a client side outgoing connection.
func NewServerConn(c net.Conn, targetAddr string, password string) Conn {
	return &serverConn{c, targetAddr, kdf(password, 32)}
}

// NewSocksConn wrap a client side incoming socks connection.
func NewSocksConn(c net.Conn) Conn {
	return &socksConn{c}
}

// NewHTTPConn wrap a client side incoming http connection.
func NewHTTPConn(c net.Conn) Conn {
	return &httpConn{Conn: c}
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

// Handshake do handshake to app.
func (c *httpConn) Handshake() (string, error) {
	return c.handshake()
}

// Relay copies between left and right bidirectionally.
func Relay(left, right net.Conn) error {
	ch := make(chan error, 2)

	go func() {
		_, err := io.Copy(right, left)
		ch <- err
		right.SetReadDeadline(time.Now()) // unblock read on right
	}()

	_, err := io.Copy(left, right)
	ch <- err
	left.SetReadDeadline(time.Now()) // unblock read on left

	// the first err is relay error reason.
	err = <-ch
	// wait goroutine done
	<-ch
	return err
}
