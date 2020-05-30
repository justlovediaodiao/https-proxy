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

// handshake do server side handshake to client.
// Read client handshake http request and verify authorization.
// Response http 200 if success else other status code.
// Return target address that client want to connect to.
func (c *clientConn) handshake() (string, error) {
	req, err := http.ReadRequest(bufio.NewReader(c.Conn))
	if err != nil {
		httpResponse(c.Conn, http400)
		return "", err
	}
	defer req.Body.Close()
	if req.Method != "GET" || req.URL.Path != "/" {
		httpResponse(c.Conn, http403)
		return "", err
	}
	addr, ok := verifyAuthQuery(req.URL.Query(), c.password)
	if !ok {
		httpResponse(c.Conn, http403)
		return "", errors.New(http403)
	}
	httpResponse(c.Conn, http200)
	return addr, nil
}

// handshake do client side handshake to server.
// Send handshake http request to sever and read sever response.
// Success if sever response http 200.
func (c *serverConn) handshake() error {
	var q = getAuthQuery(c.targetAddr, c.password)
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
func (c *httpConn) handshake() (string, error) {
	var bufConn = bufio.NewReader(c.Conn)
	req, err := http.ReadRequest(bufConn)
	if err != nil {
		return "", err
	}
	if req.Method == "CONNECT" { // tunnel mode, for https
		c.isTunnel = true
		req.Body.Close()
		if err = httpResponse(c.Conn, http200); err != nil {
			return "", err
		}
		return joinHostPort(req.URL.Host, "80"), nil
	}
	// proxy mode, for http
	c.bufConn = bufConn
	c.request = newRequestReader(req)
	return joinHostPort(req.URL.Host, "80"), nil
}

// Read reads data from the connection.
func (c *httpConn) Read(b []byte) (int, error) {
	if c.isTunnel { // tunnnel mode just relay after handshake
		return c.Conn.Read(b)
	}
READ:
	// proxy mode should forward http request to server
	if c.request == nil {
		req, err := http.ReadRequest(c.bufConn)
		if err != nil {
			return 0, err
		}
		c.request = newRequestReader(req)
	}
	n, err := c.request.Read(b)
	if err == io.EOF { // EOF, request ended
		c.request = nil
		err = nil
		goto READ
	}
	return n, err
}

// requestReader trun http.Request to stream used to forward to remote.
// Read call will automatically close req.Body when read to EOF or error.
type requestReader struct {
	req       *http.Request
	reqReader io.Reader
	eof       bool
}

// Read read data as much as possiable until full or EOF or error.
func (r *requestReader) Read(b []byte) (n int, err error) {
	if r.eof {
		err = io.EOF
		return
	}
	for n < len(b) && err == nil {
		var nn int
		nn, err = r.reqReader.Read(b[n:])
		n += nn
	}
	if err != nil {
		r.req.Body.Close() // must close req.Body
	}
	if n > 0 && err == io.EOF { // should not return eof if n > 0
		r.eof = true
		err = nil
	}
	return
}

// newRequestReader return requestReader.
func newRequestReader(req *http.Request) io.Reader {
	var rs = make([]io.Reader, 0, len(req.Header)+3) // request line + header lines + \r\n + body. assume each header appears once.
	var reqLine = fmt.Sprintf("%s %s HTTP/1.1\r\n", req.Method, req.URL.RequestURI())
	rs = append(rs, strings.NewReader(reqLine))
	for k, vs := range req.Header {
		// remove hop-by-hop headers, not sure, fuck http specification.
		switch k {
		case "Transfer-Encoding": // request body maybe chuncked, but forwarding to remote is not.
		case "Proxy-Authenticate":
		case "Proxy-Authorization":
		case "Connection":
		case "Trailer":
		case "TE":
		case "Upgrade": // maybe websocket, donot support.
			continue
		case "Proxy-Connection":
			k = "Connection"
		}
		for _, v := range vs {
			var header = fmt.Sprintf("%s: %s\r\n", k, v)
			rs = append(rs, strings.NewReader(header), req.Body)
		}
	}
	// set Host header and \r\n
	rs = append(rs, strings.NewReader(fmt.Sprintf("Host: %s\r\n\r\n", req.Host)))
	return &requestReader{req, io.MultiReader(rs...), false}
}
