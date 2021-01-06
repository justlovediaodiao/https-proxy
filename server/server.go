package server

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/justlovediaodiao/https-proxy/proxy"
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
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}
		go handleConn(conn, config.Password)
	}
}

func handleConn(conn net.Conn, password string) {
	defer conn.Close()
	// tls handshake
	var err = conn.(*tls.Conn).Handshake()
	if err != nil {
		log.Printf("tls handshake error: %v", err)
		return
	}
	// proxy client http handshake
	var cc = proxy.NewClientConn(conn, password)
	addr, err := cc.Handshake()
	if err != nil {
		log.Printf("http handshake error: %v", err)
		return
	}
	// connect to target server and relay stream
	rc, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("failed to dial %v: %v", addr, err)
		return
	}
	defer rc.Close()

	log.Printf("%v <----> %v", cc.RemoteAddr().String(), addr)
	err = proxy.Relay(cc, rc)
	if err != nil {
		log.Printf("relay error: %v", err)
	}
}
