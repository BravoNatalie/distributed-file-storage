package p2p

// Peer is a interface that represent the remote node.
type Peer interface {
	Close() error
}

// Transport is anything that handlers the comunication
// between the nodes. This can be of the form (TCP, UDP, gRPC, ...)
type Transport interface {
	ListenAndAccept() error
	Consume() <- chan RPC
}

