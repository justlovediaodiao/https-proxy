# https-proxy

[![Build](https://github.com/justlovediaodiao/https-proxy/workflows/Build/badge.svg)](https://github.com/justlovediaodiao/https-proxy/actions?query=workflow%3ABuild)
[![Go version](https://img.shields.io/github/go-mod/go-version/justlovediaodiao/https-proxy)](https://golang.org/)

HTTPS proxy is a tcp/udp proxy. It transfers proxy data over HTTPS.

### Why HTTPS

- HTTPS is widely used and its security is verified enough.  
- Make the proxy looks like a HTTPS communication to avoid detection and being blocked.

### Certificate

- If you have a domain and a trusted certificate signed by CA, you can use it directly.

- Or use a self-signed certificate and let client trust it. The repositorie provides a `cert` command to generate certificates.

1. install `openssl` if not.
3. run `cert -ip <server ip>` or `cert -host <server domain>` to generate a certificate. You will get `hp.key` and `hp.crt` files. Do not leak out `hp.key`.

### Usage

**server:**

```
hpserver -l :443 -cert hp.crt -key hp.key -password F09a5SZbhJfzp5GI
```

It will start a https server on :443 and use `hp.crt` as certificate with key `hp.key`.

- l: server listening address, default is `:443`.
- cert: tls certificate file path.
- key: tls certificate key file path.
- password: password used to verify client.

**client:**

```
hpclient -l 127.0.0.1:1080 -server 59.24.3.174:443 -cert hp.crt -password F09a5SZbhJfzp5GI
```

It will start a socks5 proxy listening tcp/udp on `127.0.0.1:1080` and proxy to `59.24.3.174:443`.  
`hp.crt` is trusted root CA.

- l: local listening address, default is `:1080`.
- http: listening for http proxy, not socks, which donot support udp.
- server: server address.
- cert: root certificate file path, used to verify server's certificate. optional, needed when using a self-signed certificate. 
- password: password used for authorization.

### Protocol

```
[tls handhake] [encrypted payload]
```

- tls handshake: On Handshake, client and server will negotiate encryption method and encryption key used for encrypting payload. see [tls handshake](https://en.wikipedia.org/wiki/Transport_Layer_Security#TLS_handshake).
- encrypted payload:

```
[http handshake] [tcp data]
```

- http handshake: 

Client send a http request to server:
```
GET /?target=github.com:443&time=1590411634&sig=c2208abde9668e8e9815c3690855edd1e63abeac
```

- method: Must be `GET`.
- path: Must be `/`.
- network: udp or tcp.
- target: Target address with port. ipv4 or ipv6 or domain.
- time: Current unix timestamp that is accurate to a second. No more than 2 minutes compared to server time.
- sig: Signature. `HMAC(msg, key, sha1)`:
```
msg: network + target + time string, for example: tcpgithub.com:4431590411634
key: 32-bytes, derivation from password. following EVP_BytesToKey(3) in OpenSSL.
sha1: the SHA1 hash algorithm
```

If authorization success, sever must response http status code 200. Other response codes are considered failures.

```
HTTP/1.1 200 OK
```

Don't return too many http error codes. The client doesn't care about this. Instead, it gives the chance to detect whether it is a proxy service.

- proxy data: Real transfer data.


#### UDP

UDP packets are transfered over tcp. see [udp-over-tcp](https://github.com/justlovediaodiao/udp-over-tcp).
