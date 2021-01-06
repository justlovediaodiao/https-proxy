package proxy

import (
	"net"
	"net/url"
	"testing"
)

func TestAuth(t *testing.T) {
	var addr = "github.com"
	var key = kdf("test", 32)
	var query = getAuthQuery(addr, key)
	q, err := url.ParseQuery(query)
	if err != nil {
		t.Error(err)
	}
	addr2, ok := verifyAuthQuery(q, key)
	if !ok || addr != addr2 {
		t.Fail()
	}
}

func TestHttpConn(t *testing.T) {
	var data = []string{
		"GET http://github.com/index?name=abc HTTP/1.1\r\nProxy-Connection: keep-alive\r\nUser-Agent: golang\r\n\r\n",
		"POST http://github.com/post?name=def HTTP/1.1\r\nContent-Length: 11\r\n\r\nhello world",
	}
	var laddr = "127.0.0.1:1080"
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		t.Error(err)
	}

	go func() {
		c, err := net.Dial("tcp", laddr)
		if err != nil {
			t.Error(err)
		}
		for _, d := range data {
			_, err := c.Write([]byte(d))
			if err != nil {
				t.Error(err)
			}
		}
		c.Close()
	}()

	conn, err := l.Accept()
	if err != nil {
		t.Error(err)
	}

	var c = NewHTTPConn(conn)
	addr, err := c.Handshake()
	if err != nil {
		t.Error(err)
	}
	t.Log(addr)

	var b = make([]byte, 1024)
	for i := 0; i < len(data); i++ {
		n, err := c.Read(b)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%d: %s", i, b[:n])
	}
}
