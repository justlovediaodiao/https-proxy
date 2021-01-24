package main

import (
	"flag"
	"fmt"

	"github.com/justlovediaodiao/https-proxy/client"
)

func main() {
	var c = client.Config{}
	var http bool
	flag.StringVar(&c.Listen, "l", ":1080", "listen address")
	flag.BoolVar(&http, "http", false, "listen for http proxy")
	flag.StringVar(&c.Server, "server", "", "server address")
	flag.StringVar(&c.Cert, "cert", "", "tls certificate cert file, optional")
	flag.StringVar(&c.Password, "password", "", "password")
	flag.Parse()
	if c.Server == "" || c.Password == "" {
		flag.Usage()
		return
	}
	if http {
		c.Protocol = "http"
	} else {
		c.Protocol = "socks"
	}
	var err = client.Start(&c)
	if err != nil {
		fmt.Printf("failed to start client: %v", err)
	}
}
