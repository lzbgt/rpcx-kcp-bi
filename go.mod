module github.com/lzbgt/rpcx-kcp-bi

go 1.15

replace google.golang.org/grpc => google.golang.org/grpc v1.29.0

require (
	github.com/rpcxio/rpcx-examples v1.1.6
	github.com/smallnest/rpcx v0.0.0-20201112102542-4308d27f440e
	github.com/xtaci/kcp-go v5.4.20+incompatible
	golang.org/x/crypto v0.0.0-20201112155050-0c6587e931a9
)
