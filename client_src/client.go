//go run -tags kcp client.go
package main

import (
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	example "github.com/rpcxio/rpcx-examples"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
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
	bc, _ := kcp.NewAESBlockCrypt(pass)
	option := client.DefaultOption
	option.Block = bc
	option.Heartbeat = false
	option.HeartbeatInterval = time.Second
	option.ConnectTimeout = time.Second * 5
	option.IdleTimeout = 0 // never timeout
	option.Retries = 0

	d := client.NewPeer2PeerDiscovery("kcp@"+*addr, "")
	//xclient := client.NewXClient("Arith", client.Failtry, client.RoundRobin, d, option)
	ch := make(chan *protocol.Message)
	xclient := client.NewBidirectionalXClient("Arith", client.Failfast, client.RoundRobin, d, option, ch)

	// plugin
	cs := &ConfigUDPSession{}
	pc := client.NewPluginContainer()
	pc.Add(cs)
	xclient.SetPlugins(pc)

	args := &example.Args{
		A: 10,
		B: 20,
	}

	start := time.Now()
	for i := 0; i < 1; i++ {
		reply := &example.Reply{}
		err := xclient.Call(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Fatalf("failed to call: %v", err)
		}
		//log.Printf("%d * %d = %d", args.A, args.B, reply.C)
	}
	dur := time.Since(start)
	qps := 1 * 1000 / int(dur/time.Millisecond)
	fmt.Printf("qps: %d call/s", qps)

	go func() {
		for {
			reply := &example.Reply{}
			fmt.Println("sending beat")
			ctx, fnCancel := context.WithTimeout(context.Background(), time.Second*2)
			err := xclient.Call(ctx, "Mul", args, reply)
			fmt.Println("sent beat")
			if err != nil {
				fmt.Printf("failed from server: %v\n", err)
				xclient = client.NewBidirectionalXClient("Arith", client.Failfast, client.RoundRobin, d, option, ch)
			} else {
				fmt.Println("go beat")
			}
			fnCancel()

			time.Sleep(1 * time.Second)
		}
	}()

	for msg := range ch {
		fmt.Printf("receive msg from server: %s\n", msg.Payload)
	}

	xclient.Close()
}

type ConfigUDPSession struct{}

func (p *ConfigUDPSession) ConnCreated(conn net.Conn) (net.Conn, error) {
	session, ok := conn.(*kcp.UDPSession)
	if !ok {
		return conn, nil
	}

	session.SetACKNoDelay(true)
	session.SetStreamMode(true)

	clientConn = conn
	return conn, nil
}
