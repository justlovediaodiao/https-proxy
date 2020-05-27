package proxy

import (
	"errors"
	"io"
	"net"
	"strconv"
)

// maxAddrLen is the max size of socks address host in bytes.
const maxAddrLen = 1 + 1 + 255

// cmdConnect is the value of cmd field on handshake.
const cmdConnect = 1

// socksVer is socks version.
const socksVer = 5

// socks address type. see RFC 1928.
const (
	atypIPv4       = 1
	atypDomainName = 3
	AtypIPv6       = 4
)

// readAddr read socks addr defined in RFC 1928.
func (c *socksConn) readAddr(buf []byte) (string, error) {
	_, err := io.ReadFull(c.Conn, buf[:1]) // read 1st byte for address type
	if err != nil {
		return "", err
	}
	var host string
	switch buf[0] {
	case atypDomainName:
		_, err = io.ReadFull(c.Conn, buf[1:2]) // read 2nd byte for domain length
		if err != nil {
			return "", err
		}
		_, err = io.ReadFull(c.Conn, buf[2:2+int(buf[1])])
		host = string(buf[2 : 2+int(buf[1])])
	case atypIPv4:
		_, err = io.ReadFull(c.Conn, buf[1:1+4])
		host = net.IP(buf[1 : 1+4]).String()
	case AtypIPv6:
		_, err = io.ReadFull(c.Conn, buf[1:1+16])
		host = net.IP(buf[1 : 1+16]).String()
	default:
		err = errors.New("error socks address")
	}
	if err != nil {
		return "", err
	}
	// read 2-byte port
	_, err = io.ReadFull(c.Conn, buf[:2])
	var port = strconv.Itoa((int(buf[0]) << 8) | int(buf[1]))
	return net.JoinHostPort(host, port), nil
}

// handshake do proxy side socks5 handshake. Return target address that app want to connect to.
func (c *socksConn) handshake() (string, error) {
	buf := make([]byte, maxAddrLen)
	// read VER, NMETHODS, METHODS
	if _, err := io.ReadFull(c.Conn, buf[:2]); err != nil {
		return "", err
	}
	ver := buf[0]
	if ver != socksVer {
		return "", errors.New("not a socks5 protocol")
	}
	nmethods := buf[1]
	if _, err := io.ReadFull(c.Conn, buf[:nmethods]); err != nil {
		return "", err
	}
	// write VER METHOD
	if _, err := c.Conn.Write([]byte{socksVer, 0}); err != nil {
		return "", err
	}
	// read VER CMD RSV ATYP DST.ADDR DST.PORT
	if _, err := io.ReadFull(c.Conn, buf[:3]); err != nil {
		return "", err
	}
	cmd := buf[1]
	if cmd != cmdConnect {
		return "", errors.New("unsupported socks command")
	}
	addr, err := c.readAddr(buf)
	if err != nil {
		return "", err
	}
	_, err = c.Conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}) // SOCKS v5, reply succeeded

	return addr, err // skip VER, CMD, RSV fields
}
