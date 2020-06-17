package proxy

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// sign calc HMAC(msg, key, sha1)
func sign(msg []byte, key []byte) []byte {
	var h = hmac.New(sha1.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

// getAuthQuery return http request query string used for authorization.
func getAuthQuery(targetAddr string, password string) string {
	var ts = fmt.Sprintf("%d", time.Now().Unix())
	var msg = fmt.Sprintf("%s%s", targetAddr, ts)
	var sig = fmt.Sprintf("%x", sign([]byte(msg), []byte(password)))
	var q = url.Values{
		"time":   []string{ts},
		"target": []string{targetAddr},
		"sig":    []string{sig},
	}
	return q.Encode()
}

// verifyAuthQuery verify http request query string and return target address.
func verifyAuthQuery(q url.Values, password string) (string, bool) {
	var targetAddr = q.Get("target")
	var ts = q.Get("time")
	var sig = q.Get("sig")
	if targetAddr == "" || ts == "" || sig == "" {
		return "", false
	}
	t, err := strconv.Atoi(ts)
	if err != nil {
		return "", false
	}
	var diff = time.Now().Unix() - int64(t)
	if diff < -120 || diff > 120 {
		return "", false
	}
	var msg = fmt.Sprintf("%s%s", targetAddr, ts)
	var sig2 = sign([]byte(msg), []byte(password))
	if fmt.Sprintf("%x", sig2) != sig {
		return "", false
	}
	return targetAddr, true
}
