package server

// Config config.
type Config struct {
	// listen address
	Listen string
	// pem cert file
	Cert string
	// pem cert private key
	Key string
	// predefined password with client
	Password string
}
