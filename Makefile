all: server client
.phony: all clean
server: server.go
	go build -tags kcp -o server server.go
client: client.go
	go build -tags kcp -o client client.go

clean:
	rm -fr server client
