package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net"
	"os"

	"github.com/justlovediaodiao/https-proxy/proxy"
	uot "github.com/justlovediaodiao/udp-over-tcp"
)

var tlsConfig = &tls.Config{MinVersion: tls.VersionTLS13}

// Start start client service.
func Start(config *Config) error {
	if config.Cert != "" {
		if err := configRootCA(config.Cert); err != nil {
			return err
		}
	}

	if config.Protocol == "socks" {
		conn, err := net.ListenPacket("udp", config.Listen)
		if err != nil {
			return err
		}
		go startUDP(conn, config.Server, config.Password)
	}

	l, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return err
	}
	log.Printf("listening on %s for %s", l.Addr().String(), config.Protocol)

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
		go handleConn(conn, config.Server, config.Password, config.Protocol)
	}
}

func configRootCA(cert string) error {
	var pool = x509.NewCertPool()
	b, err := os.ReadFile(cert)
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

func handleConn(conn net.Conn, server string, password string, protocol string) {
	defer conn.Close()
	// client proxy handshake
	var cc proxy.Conn
	if protocol == "socks" {
		cc = proxy.NewSocksConn(conn)
	} else {
		cc = proxy.NewHTTPConn(conn)
	}
	addr, err := cc.Handshake(nil)
	if err != nil {
		// udp associate, keep connection and finally close.
		if err == proxy.ErrUDPAssociate {
			log.Printf("%s udp associate", addr.String())
			conn.Read(make([]byte, 1))
			return
		}
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
	var sc = proxy.NewOutConn(rc, password)
	_, err = sc.Handshake(addr)
	if err != nil {
		log.Printf("http handshake error: %v", err)
		return
	}

	log.Printf("%v <---> %v <---> %v", cc.RemoteAddr().String(), sc.RemoteAddr().String(), addr)
	err = proxy.Relay(cc, sc)
	if err != nil {
		log.Printf("relay error: %v", err)
	}
}

func startUDP(conn net.PacketConn, server string, password string) {
	h := uot.Client{
		Dialer: func(addr string) (uot.Conn, error) {
			rc, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				log.Printf("failed to dial %v: %v", addr, err)
				return nil, err
			}
			err = rc.Handshake()
			if err != nil {
				log.Printf("tls handshake error: %v", err)
				return nil, err
			}
			cc := proxy.NewOutConn(rc, password)
			return proxy.UDPOverTCP(cc), nil
		},
		Logf: log.Printf,
	}
	h.Serve(proxy.NewSocksUDPConn(conn), server)
}
