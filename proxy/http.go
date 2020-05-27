package proxy

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// httpReader read http request line from net.Conn
type httpReader struct {
	Conn  net.Conn
	buf   []byte
	last  byte
	start int
	end   int
}

// maxLineLen is max http request line length
const maxLineLen = 8192

// line break
const (
	cr byte = 13 // "\r"
	lf byte = 10 // "\n"
)

// http status code and description.
const (
	http200 = "200 OK"
	http403 = "403 Forbidden"
	http400 = "400 Bad Request"
)

// ReadLine read a http request line endswith \r\n
// The return []byte is a slice of inner bufffer of httpReader.
// It may be changed after next ReadLine call.
func (r *httpReader) ReadLine() ([]byte, error) {
	if r.buf == nil {
		r.buf = make([]byte, maxLineLen)
	} else {
		// read from buffer
		for i, v := range r.buf[r.start:r.end] {
			if r.last == cr && v == lf {
				var result = r.buf[r.start : r.start+i+1]
				r.start += i + 1
				r.last = v
				return result, nil
			}
			r.last = v
		}
		// if not line end, copy data to buffer start, then read from io.
		copy(r.buf, r.buf[r.start:r.end])
		r.end = r.end - r.start
		r.start = 0
	}
	for {
		var start = r.end
		n, err := r.Conn.Read(r.buf[start:])
		if err != nil {
			return nil, err
		}
		r.end += n
		for i, v := range r.buf[start:r.end] {
			if r.last == cr && v == lf {
				var result = r.buf[r.start : start+i+1]
				r.start = start + i + 1
				r.last = v
				return result, nil
			}
			r.last = v
		}
		if r.end >= maxLineLen {
			return nil, errors.New("over max request line length")
		}
	}
}

// ReadToEnd read lines until an empty line which is \r\n
func (r *httpReader) ReadToEnd() error {
	for {
		line, err := r.ReadLine()
		if err != nil {
			return err
		}
		// \r\n
		if len(line) == 2 {
			return nil
		}
	}
}

// httpResponse write http response to client
func (c *clientConn) httpResponse(status string) error {
	var line = fmt.Sprintf("HTTP/1.1 %s\r\n\r\n", status)
	_, err := c.Conn.Write([]byte(line))
	return err
}

// handshake do server side handshake to client.
// Read client handshake http request and verify authorization.
// Response http 200 if success else other status code.
// Return target address that client want to connect to.
func (c *clientConn) handshake() (string, error) {
	// GET /?target=&time=&sig= HTTP/1.1
	var r = httpReader{Conn: c.Conn}
	b, err := r.ReadLine()
	if err != nil {
		c.httpResponse(http400)
		return "", errors.New(http400)
	}
	var arr = strings.Split(string(b), " ")
	if len(arr) != 3 || arr[2] != "HTTP/1.1\r\n" {
		c.httpResponse(http400)
		return "", errors.New(http400)
	}
	if arr[0] != "GET" {
		c.httpResponse(http403)
		return "", errors.New(http403)
	}
	targetAddr, ok := verifyUriSig(arr[1], c.password)
	if !ok {
		c.httpResponse(http403)
		return "", errors.New(http403)
	}
	c.httpResponse(http200)
	return targetAddr, nil
}

// handshake do client side handshake to server.
// Send handshake http request to sever and read sever response.
// Success if sever response http 200.
func (c *serverConn) handshake() error {
	var uri = getSignedUri(c.targetAddr, c.password)
	var line = fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", uri)
	c.Conn.Write([]byte(line))
	// HTTP/1.1 200 OK
	var r = httpReader{Conn: c.Conn}
	b, err := r.ReadLine()
	if err != nil {
		return err
	}
	var arr = strings.SplitN(string(b), " ", 3)
	if len(arr) != 3 || arr[0] != "HTTP/1.1" {
		return errors.New(http400)
	}
	code, err := strconv.Atoi(arr[1])
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("hadnshake error, code: %d", code)
	}
	return r.ReadToEnd()
}
