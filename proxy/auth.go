package proxy

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

// key-derivation function from original Shadowsocks
func kdf(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}

// sign calc HMAC(msg, key, sha1)
func sign(msg []byte, key []byte) []byte {
	var h = hmac.New(sha1.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

// getAuthQuery return http request query string used for authorization.
func getAuthQuery(targetAddr net.Addr, key []byte) string {
	var ts = fmt.Sprintf("%d", time.Now().Unix())
	var msg = fmt.Sprintf("%s%s%s", targetAddr.Network(), targetAddr.String(), ts)
	var sig = fmt.Sprintf("%x", sign([]byte(msg), key))
	var q = url.Values{
		"time":    []string{ts},
		"network": []string{targetAddr.Network()},
		"target":  []string{targetAddr.String()},
		"sig":     []string{sig},
	}
	return q.Encode()
}

// verifyAuthQuery verify http request query string and return target address.
func verifyAuthQuery(q url.Values, key []byte) (net.Addr, bool) {
	var network = q.Get("network")
	var target = q.Get("target")
	var ts = q.Get("time")
	var sig = q.Get("sig")
	if network == "" || target == "" || ts == "" || sig == "" {
		return nil, false
	}
	t, err := strconv.Atoi(ts)
	if err != nil {
		return nil, false
	}
	var diff = time.Now().Unix() - int64(t)
	if diff < -120 || diff > 120 {
		return nil, false
	}
	var msg = fmt.Sprintf("%s%s%s", network, target, ts)
	var sig2 = sign([]byte(msg), key)
	if fmt.Sprintf("%x", sig2) != sig {
		return nil, false
	}
	return targetAddr{network, target}, true
}
