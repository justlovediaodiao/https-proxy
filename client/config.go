package client

type Config struct {
	Listen   string
	Socks    bool
	Http     bool
	Server   string
	Cert     string
	Password string
}
