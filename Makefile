all: server client
.phony: all clean
server: server.go
	go build -tags kcp -o server server.go
client: client_src/client.go
	go build -tags kcp -o client client_src/client.go

clean:
	rm -fr server client_src/client
