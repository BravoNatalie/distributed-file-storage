package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPer represents the remote node over a TCP connection.
type TCPPeer struct {
	// conn is the underlying connection of the peer
	conn	net.Conn

	// True if the connection was initiated by the peer (dialed); 
	// False if the connection was accepted from a peer.
	outbound bool
}

func NewTCPPeer(conn net.Conn, isOutbound bool) *TCPPeer{
	return &TCPPeer{
		conn: conn,
		outbound: isOutbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}
 
func (t *TCPTransport) ListenAndAccept() error { 
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		fmt.Printf("new incoming connection: %v\n", conn)
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn){
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error: %s\n", err)
		return 
	}

	rpc := &RPC{}
	for {	
		if err := t.Decoder.Decode(conn, rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}

		rpc.From = conn.RemoteAddr()

		fmt.Printf("message: %+v\n", rpc)
	}

}