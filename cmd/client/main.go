package main

import (
	"flag"
	"fmt"

	"github.com/justlovediaodiao/https-proxy/client"
)

func main() {
	var c = client.Config{}
	flag.StringVar(&c.Listen, "l", ":1080", "listen address")
	flag.BoolVar(&c.Socks, "socks", false, "listen for socks5 proxy, which is default")
	flag.BoolVar(&c.HTTP, "http", false, "listen for http proxy")
	flag.StringVar(&c.Server, "server", "", "server address")
	flag.StringVar(&c.Cert, "cert", "", "tls certificate cert file, optional")
	flag.StringVar(&c.Password, "password", "", "password")
	flag.Parse()
	if c.Server == "" || c.Password == "" {
		flag.Usage()
		return
	}
	if c.Socks && c.HTTP {
		c.HTTP = false
	} else if !c.Socks && !c.HTTP {
		c.Socks = true
	}
	var err = client.Start(&c)
	if err != nil {
		fmt.Printf("failed to start client: %v", err)
	}
}
