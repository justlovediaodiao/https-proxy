package server

import (
	"crypto/tls"
	"errors"
	"log"
	"net"

	"github.com/justlovediaodiao/https-proxy/proxy"
	uot "github.com/justlovediaodiao/udp-over-tcp"
)

// Start start server service.
func Start(config *Config) error {
	cer, err := tls.LoadX509KeyPair(config.Cert, config.Key)
	if err != nil {
		return err
	}
	var c = &tls.Config{Certificates: []tls.Certificate{cer}, MinVersion: tls.VersionTLS13}
	l, err := tls.Listen("tcp", config.Listen, c)
	if err != nil {
		return err
	}
	log.Printf("listening on %v for https", l.Addr().String())
	udpHandler := uot.Server{Logf: log.Printf}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			if errors.Is(err, net.ErrClosed) {
				log.Printf("tcp listener closed, exit")
				return nil
			}
			continue
		}
		go handleConn(conn, config.Password, udpHandler)
	}
}

func handleConn(conn net.Conn, password string, udpHandler uot.Server) {
	defer conn.Close()
	// tls handshake
	var err = conn.(*tls.Conn).Handshake()
	if err != nil {
		log.Printf("tls handshake error: %v", err)
		return
	}
	// proxy client http handshake
	var cc = proxy.NewInConn(conn, password)
	addr, err := cc.Handshake(nil)
	if err != nil {
		log.Printf("http handshake error: %v", err)
		return
	}
	if addr.Network() == "tcp" {
		handleTCP(cc, addr)
	} else if addr.Network() == "udp" {
		cc = proxy.UDPOverTCP(cc)
		// should not handshake again
		udpHandler.Serve(noHandshake{cc, addr})
	} else {
		log.Printf("unkown network: %s", addr.Network())
	}
}

func handleTCP(conn proxy.Conn, addr net.Addr) {
	// connect to target server and relay stream
	rc, err := net.Dial("tcp", addr.String())
	if err != nil {
		log.Printf("failed to dial %v: %v", addr, err)
		return
	}
	defer rc.Close()

	log.Printf("%v <---> %v", conn.RemoteAddr().String(), addr)
	err = proxy.Relay(conn, rc)
	if err != nil {
		log.Printf("relay error: %v", err)
	}
}

type noHandshake struct {
	proxy.Conn
	addr net.Addr
}

func (c noHandshake) Handshake(net.Addr) (net.Addr, error) {
	return c.addr, nil
}
