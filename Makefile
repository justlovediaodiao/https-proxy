BIN=bin
MODULE=github.com/justlovediaodiao/https-proxy

.PHONY: server client cert

all: server client cert

server:
	go build -o $(BIN)/hpserver $(MODULE)/cmd/server

client:
	go build -o $(BIN)/hpclient $(MODULE)/cmd/client

cert:
	go build -o $(BIN)/cert $(MODULE)/cmd/cert

test: server client cert
	cd $(BIN) && ./cert -ip 127.0.0.1
	echo "./hpserver -cert hp.crt -key hp.key -password test" > $(BIN)/server.sh
	echo "./hpclient -cert hp.crt -server 127.0.0.1:443 -password test -http" > $(BIN)/client.sh
	chmod u+x $(BIN)/server.sh
	chmod u+x $(BIN)/client.sh

clean:
	rm -r $(BIN)
