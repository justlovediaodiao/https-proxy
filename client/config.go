package client

// Config config.
type Config struct {
	// listen address
	Listen string
	// socks or http
	Protocol string
	// server address
	Server string
	// pem cert content, optional
	Cert string
	// predefined password with server
	Password string
}
