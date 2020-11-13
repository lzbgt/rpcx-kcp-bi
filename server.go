//go run -tags kcp server.go
package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"net"
	"time"

	example "github.com/rpcxio/rpcx-examples"
	"github.com/smallnest/rpcx/server"
	kcp "github.com/xtaci/kcp-go"
	"golang.org/x/crypto/pbkdf2"
)

var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

const cryptKey = "rpcx-key"
const cryptSalt = "rpcx-salt"

var clientConn net.Conn

func main() {
	flag.Parse()

	pass := pbkdf2.Key([]byte(cryptKey), []byte(cryptSalt), 4096, 32, sha1.New)
	bc, err := kcp.NewAESBlockCrypt(pass)
	if err != nil {
		panic(err)
	}

	s := server.NewServer(server.WithBlockCrypt(bc))
	s.RegisterName("Arith", new(example.Arith), "")

	cs := &ConfigUDPSession{}
	s.Plugins.Add(cs)

	go s.Serve("kcp", *addr)
	for {
		if clientConn != nil {
			err := s.SendMessage(clientConn, "test_service_path", "test_service_method", nil, []byte("abcde"))
			if err != nil {
				fmt.Printf("failed to send messsage to %s: %v\n", clientConn.RemoteAddr().String(), err)
				clientConn = nil
			}
		} else {
			fmt.Println("nil conn")
		}

		time.Sleep(1 * time.Second)

	}
}

type ConfigUDPSession struct{}

func (p *ConfigUDPSession) HandleConnAccept(conn net.Conn) (net.Conn, bool) {
	session, ok := conn.(*kcp.UDPSession)
	if !ok {
		return conn, true
	}

	session.SetACKNoDelay(true)
	session.SetStreamMode(true)
	clientConn = conn
	return conn, true
}
