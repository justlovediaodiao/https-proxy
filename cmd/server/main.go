package main

import (
	"flag"
	"fmt"

	"github.com/justlovediaodiao/https-proxy/server"
)

func main() {
	var c = server.Config{}
	flag.StringVar(&c.Listen, "l", ":443", "listen address")
	flag.StringVar(&c.Cert, "cert", "", "tls certificate cert file")
	flag.StringVar(&c.Key, "key", "", "tls certificate key file")
	flag.StringVar(&c.Password, "password", "", "password")
	flag.Parse()
	if c.Cert == "" || c.Key == "" || c.Password == "" {
		flag.Usage()
		return
	}
	var err = server.Start(&c)
	if err != nil {
		fmt.Printf("failed to start server: %v", err)
	}
}
