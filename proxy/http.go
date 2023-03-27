package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// http status code and description.
const (
	http200 = "200 OK"
	http403 = "403 Forbidden"
	http400 = "400 Bad Request"
)

// httpResponse write http response to client
func httpResponse(conn net.Conn, status string) error {
	var line = fmt.Sprintf("HTTP/1.1 %s\r\n\r\n", status)
	_, err := conn.Write([]byte(line))
	return err
}

// joinHostPort join port to host if host contains no port
func joinHostPort(host string, port string) string {
	if strings.LastIndexByte(host, ':') == -1 || strings.HasSuffix(host, "]") { // ipv6 addr [...]
		return fmt.Sprintf("%s:%s", host, port)
	}
	return host
}

// serverHandshake do server side handshake to client.
// Read client handshake http request and verify authorization.
// Response http 200 if success else other status code.
// Return target address that client want to connect to.
func (c *tcpConn) serverHandshake() (net.Addr, error) {
	req, err := http.ReadRequest(bufio.NewReader(c.Conn))
	if err != nil {
		httpResponse(c.Conn, http400)
		return nil, err
	}
	defer req.Body.Close()
	if req.Method != "GET" || req.URL.Path != "/" {
		httpResponse(c.Conn, http403)
		return nil, errors.New(http403)
	}
	addr, ok := verifyAuthQuery(req.URL.Query(), c.key)
	if !ok {
		httpResponse(c.Conn, http403)
		return nil, errors.New(http403)
	}
	if err = httpResponse(c.Conn, http200); err != nil {
		return nil, err
	}
	return addr, nil
}

// clientHandshake do client side handshake to server.
// Send handshake http request to sever and read sever response.
// Success if sever response http 200.
func (c *tcpConn) clientHandshake(addr net.Addr) error {
	var q = getAuthQuery(addr, c.key)
	var line = fmt.Sprintf("GET /?%s HTTP/1.1\r\n\r\n", q)
	_, err := c.Conn.Write([]byte(line))
	if err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(c.Conn), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hadnshake error, status code: %d", resp.StatusCode)
	}
	return nil
}

// handshake do proxy side http handshake to app.
// Return target address that app want to connect to.
func (c *httpConn) handshake() (net.Addr, error) {
	var bufConn = bufio.NewReader(c.Conn)
	req, err := http.ReadRequest(bufConn)
	if err != nil {
		return nil, err
	}
	if req.Method == "CONNECT" { // tunnel mode, for https
		req.Body.Close()
		if err = httpResponse(c.Conn, http200); err != nil {
			return nil, err
		}
	} else { // proxy mode, for http
		c.r = http2Tunnel(req, bufConn)
	}

	return targetAddr{"tcp", joinHostPort(req.URL.Host, "80")}, nil
}

func http2Tunnel(req *http.Request, bufConn *bufio.Reader) io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		for {
			err := req.Write(w)
			if err != nil {
				w.CloseWithError(err)
				break
			}
			req, err = http.ReadRequest(bufConn)
			if err != nil {
				w.CloseWithError(err)
				break
			}
		}
	}()
	return r
}

// Read reads data from connection.
func (c *httpConn) Read(b []byte) (int, error) {
	if c.r != nil {
		return c.r.Read(b)
	}
	return c.Conn.Read(b)
}

// Close close connection.
func (c *httpConn) Close() error {
	if c.r != nil {
		c.r.Close()
	}
	return c.Conn.Close()
}
