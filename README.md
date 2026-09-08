# https-proxy

HTTPS Proxy is a TCP/UDP proxy that transfers proxy traffic over TLS 1.3.

## Build

Go 1.20 or higher is required.

- Server:

```sh
go build -o hpserver ./cmd/server
```

- Client:

```sh
go build -o hpclient ./cmd/client
```

## Certificate

- If you have a domain and a certificate signed by a trusted CA, you can use that certificate directly.

- Alternatively, you can use a self-signed certificate and configure the client to trust it. This repository provides a `cert` command to generate one:

1. Install `openssl` if it is not already available.
2. Run `go build -o cert ./cmd/cert` to build the `cert` command.
3. Run `./cert -ip <server-ip>` or `./cert -host <server-domain>` to generate a certificate for the address used by the client.

The command creates `hp.key` and `hp.crt` in the current directory. Keep `hp.key` private.

## Usage

### Server

```sh
hpserver -l :443 -cert hp.crt -key hp.key -password F09a5SZbhJfzp5GI
```

This starts the proxy server on `:443`, using `hp.crt` and its private key `hp.key`.

- `-l`: Server listen address. Defaults to `:443`.
- `-cert`: TLS certificate file path. Required.
- `-key`: TLS certificate private key file path. Required.
- `-password`: Pre-shared password used to authenticate clients. Required.

### Client

```sh
hpclient -l 127.0.0.1:1080 -server 59.24.3.174:443 -cert hp.crt -password F09a5SZbhJfzp5GI
```

This starts a SOCKS5 proxy listening for TCP and UDP traffic on `127.0.0.1:1080`, and forwards it through the proxy server at `59.24.3.174:443`. In this example, `hp.crt` is trusted as the root certificate when verifying the server.

- `-l`: Local listen address. Defaults to `:1080`.
- `-server`: Proxy server address. Required.
- `-cert`: Root certificate file used to verify the server certificate. Optional for certificates signed by a system-trusted CA; required when using the generated self-signed certificate.
- `-password`: Pre-shared password used for authentication. Required and must match the server password.
- `-http`: Run an HTTP/HTTPS proxy instead of SOCKS5. HTTP proxy mode supports TCP only, not UDP.

## Protocol

```
[TLS handshake] [encrypted payload]
```

- TLS handshake: The client and server negotiate the encryption method and keys used to protect the payload. See [TLS handshake](https://en.wikipedia.org/wiki/Transport_Layer_Security#TLS_handshake). Both sides require TLS 1.3 or later.
- Encrypted payload:

```
[HTTP handshake] [TCP data]
```

- HTTP handshake:

The client sends an HTTP request to the server:

```http
GET /?network=tcp&target=github.com:443&time=1590411634&sig=c2208abde9668e8e9815c3690855edd1e63abeac
```

- Method: Must be `GET`.
- Path: Must be `/`.
- `network`: `tcp` or `udp`.
- `target`: Target IPv4, IPv6, or domain address, including the port.
- `time`: Current Unix timestamp in seconds. It must be within 2 minutes of the server time.
- `sig`: Hex-encoded `HMAC-SHA1(msg, key)` signature:

```
msg: network + target + time string, for example: tcpgithub.com:4431590411634
key: a 32-byte key derived from the password using the OpenSSL EVP_BytesToKey-compatible MD5 derivation
```

If authentication succeeds, the server responds with HTTP status `200`. Any other status code is considered a failure.

```http
HTTP/1.1 200 OK
```

- Proxy data: The actual proxied traffic.

### UDP

UDP packets are transferred over TCP. See [udp-over-tcp](https://github.com/justlovediaodiao/udp-over-tcp).
