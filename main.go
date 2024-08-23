package main

import (
	"log"

	"github.com/bravonatalie/distributed-file-storage/p2p"
)

// func OnPeer(peer p2p.Peer) error{
// 	fmt.Println("doing some logic with the peer outside the TCP transport")
// 	peer.Close()
// 	return nil
// }

func makeServer(listenArr string, nodes ...string) *FileService {
	tcpTransport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddr: listenArr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	})

	storageOpts := StoreOpts {
		StorageRoot: listenArr + "_network_storage",
		PathTrasnformFunc: CASPathTransformFunc,
	}

	fileServiceOpts := FileServiceOpts {
		StoreOpts: storageOpts,
		Transport: tcpTransport,
		BootstrapNodes: nodes,
	}

	s:= NewFileService(fileServiceOpts)

	tcpTransport.OnPeer = s.OnPeer // TODO: this is not good, change the structure

	return s
}

func main() {
	fs1 := makeServer(":3000")
	fs2 := makeServer(":4000", ":3000")
	
	go func(){
		log.Fatal(fs1.Start())
	}()


	fs2.Start()

}
