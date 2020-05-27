package proxy

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type testConn struct {
	data  []*strings.Reader
	index int
}

func newTestConn(args []string) *testConn {
	var data = make([]*strings.Reader, len(args))
	for i, v := range args {
		data[i] = strings.NewReader(v)
	}
	return &testConn{data, 0}
}

func (c *testConn) Read(b []byte) (int, error) {
	var r = c.data[c.index]
	n, err := r.Read(b)
	if err == io.EOF {
		c.index++
		if c.index == len(c.data) {
			return 0, io.EOF
		}
		r = c.data[c.index]
		return r.Read(b)
	}
	return n, err
}

func (c *testConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (c *testConn) Close() error {
	return nil
}

func (c *testConn) LocalAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("rcp", "127.0.0.1:3333")
	return addr
}

func (c *testConn) RemoteAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("rcp", "127.0.0.1:80")
	return addr
}

func (c *testConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *testConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *testConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestHttpReader(t *testing.T) {
	var cases = [][]string{
		{"GET / HTTP/1.1\r\n\r\n"},
		{"GET / HTTP", "/1.1\r\n\r\n"},
		{"GET / HTTP", "/1.1\r\n\r", "\n"},
		{"GET / HTTP/1.1\r\nContent-Type: application/json\r\n\r\n"},
		{"GET / HTTP/1.1\r\nContent-Type: ", "application/json\r\n\r\n"},
	}
	// ReadLine should return line endswith \r\n
	for _, c := range cases {
		var r = httpReader{Conn: newTestConn(c)}
		for i := 0; i < 2; i++ {
			b, err := r.ReadLine()
			if err != nil {
				t.Error(err)
			}
			if strings.LastIndex(string(b), "\r\n") == -1 {
				t.Errorf("%s", b)
			}
		}
	}
	// after ReadToEnd, ReadLine should return EOF
	for _, c := range cases {
		var r = httpReader{Conn: newTestConn(c)}
		var err = r.ReadToEnd()
		if err != nil {
			t.Error(err)
		}
		_, err = r.ReadLine()
		if err != io.EOF {
			t.Error(err)
		}
	}
}

func TestAuth(t *testing.T) {
	var cases = []struct {
		addr     string
		password string
	}{
		{"github.com:443", "test1"},
		{"115.243.232.121:443", "test2"},
	}
	for _, c := range cases {
		var uri = getSignedUri(c.addr, c.password)
		r, ok := verifyUriSig(uri, c.password)
		if !ok {
			t.Fail()
		}
		if r != c.addr {
			t.Fail()
		}
	}
}
