# helper, just internal use only jozsef
client:
	echo "Build client code"
	go build cmd/client/main.go -o ./bin/chatclient -ldflags "-s -w"
	chmod +x ./bin/chatclient

server:
	echo "Build server code"
	go build cmd/server/main.go -o ./bin/chatserver -ldflags "-s -w"
	chmod +x ./bin/chatserver

all: client server

.PHONY: server client