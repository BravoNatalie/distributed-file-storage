package p2p

// Peer is a interface that represent the remote node.
type Peer interface {

}

// Transport is anything that handlers the comunication
// between the nodes.
type Transport interface {
	ListenAndAccept() error
}

