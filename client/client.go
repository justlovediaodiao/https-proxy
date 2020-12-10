package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"net"

	"github.com/justlovediaodiao/https-proxy/proxy"
)

var tlsConfig = &tls.Config{MinVersion: tls.VersionTLS13}

func Start(config *Config) error {
	if config.Cert != "" {
		var err = configRootCA(config.Cert)
		if err != nil {
			return err
		}
	}
	l, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return err
	}
	var connF func(net.Conn) proxy.Conn
	var protocol string
	if config.Socks {
		connF = proxy.SocksConn
		protocol = "socks5"
	} else {
		connF = proxy.HttpConn
		protocol = "http"
	}
	log.Printf("listening on %s for %s", l.Addr().String(), protocol)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}
		go handleConn(conn, config.Server, config.Password, connF)
	}
}

func configRootCA(cert string) error {
	var pool = x509.NewCertPool()
	b, err := ioutil.ReadFile(cert)
	if err != nil {
		return err
	}
	ok := pool.AppendCertsFromPEM(b)
	if !ok {
		return errors.New("invalid certificate")
	}
	tlsConfig.RootCAs = pool
	return nil
}

func handleConn(conn net.Conn, server string, password string, connF func(net.Conn) proxy.Conn) {
	defer conn.Close()
	// client proxy handshake
	var cc = connF(conn)
	addr, err := cc.Handshake()
	if err != nil {
		log.Printf("proxy handshake error: %v", err)
		return
	}
	// connect to proxy server and do tls handshake
	rc, err := tls.Dial("tcp", server, tlsConfig)
	if err != nil {
		log.Printf("failed to dial %v: %v", server, err)
		return
	}
	defer rc.Close()

	err = rc.Handshake()
	if err != nil {
		log.Printf("tls handshake error: %v", err)
		return
	}
	// proxy server http handshake and relay stream
	var sc = proxy.ServerConn(rc, addr, password)
	_, err = sc.Handshake()
	if err != nil {
		log.Printf("http handshake error: %v", err)
		return
	}

	log.Printf("%v <----> %v <----> %v", cc.RemoteAddr().String(), sc.RemoteAddr().String(), addr)
	err = proxy.Relay(cc, sc)
	if err != nil {
		log.Printf("relay error: %v", err)
	}
}
