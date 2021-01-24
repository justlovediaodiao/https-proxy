package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	uot "github.com/justlovediaodiao/udp-over-tcp"
)

// Conn is a tcp stream connection.
type Conn interface {
	net.Conn
	// Handshake do handshake.
	// if receives nil, read addr from conn and return.
	// if receives non-nil, write addr to conn.
	Handshake(net.Addr) (net.Addr, error)
}

// tcpConn is connnection between client and server.
type tcpConn struct {
	net.Conn
	key      []byte
	isClient bool
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

type targetAddr struct {
	network string
	address string
}

func (a targetAddr) Network() string {
	return a.network
}

func (a targetAddr) String() string {
	return a.address
}

// NewInConn wrap a sever side incoming connection.
func NewInConn(c net.Conn, password string) Conn {
	return &tcpConn{
		Conn:     c,
		key:      kdf(password, 32),
		isClient: false,
	}
}

// NewOutConn wrap a client side outgoing connection.
func NewOutConn(c net.Conn, password string) Conn {
	return &tcpConn{
		Conn:     c,
		key:      kdf(password, 32),
		isClient: true,
	}
}

// UDPOverTCP wraps a conn between client and server as udp-over-tcp.
func UDPOverTCP(c Conn) Conn {
	cc, ok := c.(*tcpConn)
	if !ok {
		panic(fmt.Sprintf("error Conn type: %v", c))
	}
	return &uotConn{*cc}
}

// NewSocksConn wrap a client side incoming socks connection.
func NewSocksConn(c net.Conn) Conn {
	return &socksConn{c}
}

// NewSocksUDPConn wrap a client side incoming udp connection.
func NewSocksUDPConn(c net.PacketConn) uot.PacketConn {
	return uot.DefaultPacketConn(c)
}

// NewHTTPConn wrap a client side incoming http connection.
func NewHTTPConn(c net.Conn) Conn {
	return &httpConn{Conn: c}
}

// Handshake do handshake.
func (c *tcpConn) Handshake(addr net.Addr) (net.Addr, error) {
	if c.isClient {
		return addr, c.clientHandshake(addr)
	}
	return c.serverHandshake()
}

// Handshake do handshake to app.
func (c *socksConn) Handshake(net.Addr) (net.Addr, error) {
	return c.handshake()
}

// Handshake do handshake to app.
func (c *httpConn) Handshake(net.Addr) (net.Addr, error) {
	return c.handshake()
}

// Relay copies between left and right bidirectionally.
func Relay(left, right net.Conn) error {
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(right, left)
		done <- err
		right.SetReadDeadline(time.Now()) // unblock read on right
	}()

	_, err := io.Copy(left, right)
	left.SetReadDeadline(time.Now()) // unblock read on left

	// ignore timeout error.
	err1 := <-done
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		return err
	}
	if !errors.Is(err1, os.ErrDeadlineExceeded) {
		return err1
	}
	return nil
}
