package proxy

import (
	"errors"
	"io"
	"net"

	uot "github.com/justlovediaodiao/udp-over-tcp"
)

// socks request commands as defined in RFC 1928 section 4.
const (
	cmdConnect      = 1
	cmdBind         = 2
	cmdUDPAssociate = 3
)

// socksVer is socks version.
const socksVer = 5

// ErrUDPAssociate means the tcp connection used for udp associate.
var ErrUDPAssociate = errors.New("udp associate")

// handshake do proxy side socks5 handshake. Return target address that app want to connect to.
// if it is a udp associate cmd, return target address and ErrUDPAssociate.
func (c *socksConn) handshake() (net.Addr, error) {
	buf := make([]byte, 255)
	// read VER, NMETHODS, METHODS
	if _, err := io.ReadFull(c.Conn, buf[:2]); err != nil {
		return nil, err
	}
	ver := buf[0]
	if ver != socksVer {
		return nil, errors.New("not a socks5 protocol")
	}
	nmethods := buf[1]
	if _, err := io.ReadFull(c.Conn, buf[:nmethods]); err != nil {
		return nil, err
	}
	// write VER METHOD
	if _, err := c.Conn.Write([]byte{socksVer, 0}); err != nil {
		return nil, err
	}
	// read VER CMD RSV
	if _, err := io.ReadFull(c.Conn, buf[:3]); err != nil {
		return nil, err
	}
	cmd := buf[1]
	// read ATYP DST.ADDR DST.PORT
	addr, err := uot.ReadSocksAddr(c.Conn)
	if err != nil {
		return nil, err
	}
	switch cmd {
	case cmdUDPAssociate:
		a := uot.ParseSocksAddr(c.Conn.LocalAddr().String())
		if a == nil {
			return nil, errors.New("error socks address")
		}
		err = ErrUDPAssociate
		_, err = c.Conn.Write(append([]byte{5, 0, 0}, a...)) // SOCKS v5, reply succeeded
		if err != nil {
			return nil, err
		}
		return targetAddr{"tcp", addr.String()}, ErrUDPAssociate
	case cmdConnect:
		_, err = c.Conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}) // SOCKS v5, reply succeeded
		if err != nil {
			return nil, err
		}
		return targetAddr{"tcp", addr.String()}, nil
	default:
		return nil, errors.New("unsupported socks command")
	}
}
