package p2p

import "net"

// Peer is a interface that represent the remote node.
type Peer interface {
	net.Conn
	Send([]byte) error	
}

// Transport is anything that handlers the comunication
// between the nodes. This can be of the form (TCP, UDP, gRPC, ...)
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <- chan RPC
	Close() error
}

