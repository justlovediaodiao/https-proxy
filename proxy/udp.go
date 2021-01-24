package proxy

import (
	"errors"
	"io"

	uot "github.com/justlovediaodiao/udp-over-tcp"
)

// uotConn is udp-over-tcp conn.
type uotConn struct {
	tcpConn
}

// Read read a full udp packet, if b is shorter than packet, return error.
func (c *uotConn) Read(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, io.ErrShortBuffer
	}
	_, err := io.ReadFull(c.Conn, b[:2])
	if err != nil {
		return 0, err
	}
	n := int(b[0])<<8 | int(b[1])
	if len(b) < n {
		return 0, io.ErrShortBuffer
	}
	return io.ReadFull(c.Conn, b[:n])
}

// Write write a full udp packet, if head+b is longer than packet max size, return error.
func (c *uotConn) Write(b []byte) (int, error) {
	n := len(b)
	if n > uot.MaxPacketSize-2 {
		return 0, errors.New("over max packet size")
	}
	_, err := c.Conn.Write([]byte{byte(n >> 8), byte(n & 0x000000ff)})
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}
