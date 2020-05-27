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

clean:
	rm -rf $(BIN)
