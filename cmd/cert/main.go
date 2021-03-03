package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	openssl = "openssl.conf"
	cert    = "hp.crt"
	key     = "hp.key"
)

var conf = `
[req]
prompt = no
distinguished_name = dn
req_extensions = ext

[dn]
C = US
ST = California
L = San Francisco
O = GitHub
OU = justlovediaodiao
CN = %s

[ext]
subjectAltName = %s
`

func makeOpensslConfig() bool {
	var ip, host, cn, san string
	flag.StringVar(&ip, "ip", "", "sever ip, an ipv4 or ipv6 address.")
	flag.StringVar(&host, "host", "", "server domain, such as github.com or *.github.com")
	flag.Parse()
	if ip != "" {
		cn = ip
		san = "IP:" + ip
	} else if host != "" {
		cn = host
		san = "DNS:" + host
	} else {
		flag.Usage()
		return false
	}
	var content = fmt.Sprintf(conf, cn, san)
	var err = os.WriteFile(openssl, []byte(content), 0644)
	if err != nil {
		fmt.Printf("make openssl config file error: %v\n", err)
		return false
	}
	return true
}

func clean(filename string) {
	_, err := os.Stat(filename)
	if err == nil {
		os.Remove(filename)
	}
}

func main() {
	if exec.Command("openssl", "version").Run() != nil {
		fmt.Println("openssl is required")
		return
	}
	if !makeOpensslConfig() {
		return
	}
	defer clean(openssl)
	var err = exec.Command("openssl", "ecparam", "-name", "prime256v1", "-genkey", "-out", key).Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = exec.Command("openssl", "req", "-x509", "-new", "-nodes", "-key", key, "-sha256", "-days", "3650", "-config", openssl, "-extensions", "ext", "-out", cert).Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v, %v\n", key, cert)
}
