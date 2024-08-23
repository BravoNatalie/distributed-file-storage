package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
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

// RemoteAddr implements Peer interface
func (p *TCPPeer) RemoteAddr() net.Addr{
	return p.conn.RemoteAddr()
}

// TCPPeer implements Peer interface
func (p *TCPPeer) Close() error{
	return p.conn.Close()
}

func (p *TCPPeer) Send (b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

type TCPTransportOpts struct {
	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
	OnPeer func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcChan chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcChan: make(chan RPC),
	}
}

// Dial implements the Transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

// Consume implements the Transport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC{
	return t.rpcChan
}
 
// ListenAndAccept implements the Transport interface
func (t *TCPTransport) ListenAndAccept() error { 
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Println("TCP transport listening on port", t.ListenAddr, " ...")

	return nil
}

// Close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if errors.Is(err,net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		fmt.Printf("new incoming connection: %v\n", conn.RemoteAddr())
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		log.Printf("TCP handshake error: %s\n", err)
		return 
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			log.Printf("onPeer error: %s\n", err)
			return
		}
	}

	rpc := RPC{}
	for {	
		err = t.Decoder.Decode(conn, &rpc)

		if errors.Is(err,net.ErrClosed) {
			return
		}

		if err != nil {
			log.Printf("TCP read error: %s\n", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcChan <- rpc
	}

}