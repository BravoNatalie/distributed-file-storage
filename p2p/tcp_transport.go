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

type TCPTransport struct {
	listenAddress string
	listener net.Listener
	shakeHands HandshakeFunc
	decoder Decoder

	mu sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport{
	return &TCPTransport{
		listenAddress: listenAddr,
		shakeHands: NOPHandshakeFunc,
	}
}
 
func (t *TCPTransport) ListenAndAccept() error { 
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
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

type Temp struct {}

func (t *TCPTransport) handleConn(conn net.Conn){
	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(peer); err != nil {

	}

	msg := &Temp{}
	for {	
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
	}

}