package main

import (
	"flag"
	"log"

	"github.com/justlovediaodiao/https-proxy/client"
)

func main() {
	var c = client.Config{}
	flag.StringVar(&c.Listen, "l", ":1080", "listen address")
	flag.StringVar(&c.Server, "server", "", "server address")
	flag.StringVar(&c.Cert, "cert", "", "tls certificate cert file, optional")
	flag.StringVar(&c.Password, "password", "", "password")
	flag.Parse()
	if c.Server == "" || c.Password == "" {
		flag.Usage()
		return
	}
	var err = client.Start(&c)
	if err != nil {
		log.Printf("failed to start client: %v", err)
	}
}
