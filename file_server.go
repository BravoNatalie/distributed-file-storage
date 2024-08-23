package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/bravonatalie/distributed-file-storage/p2p"
)

type FileServiceOpts struct {
	Transport p2p.Transport
	StoreOpts
	BootstrapNodes []string
}

type FileService struct {
	FileServiceOpts

	store *Store
	quitch chan struct{}
	peerLock sync.Mutex 
	peers map[string]p2p.Peer
}

func NewFileService(opts FileServiceOpts) *FileService {
	return &FileService{
		FileServiceOpts: opts,
		store: NewStore(opts.StoreOpts),
		quitch: make(chan struct{}),
		peers: make(map[string]p2p.Peer),
	}
}

func (f *FileService) Start() error {
	if err := f.Transport.ListenAndAccept(); err != nil {
		return err
	}

	f.bootstrapNodes()
	f.listenForMessages()

	return nil
}

func (f *FileService) Stop() {
	 close(f.quitch)
}

func (f *FileService) OnPeer(p p2p.Peer) error {
	f.peerLock.Lock()
	defer f.peerLock.Unlock()

	f.peers[p.RemoteAddr().String()] = p

	log.Println("connected with node ", p.RemoteAddr())

	return nil
}

func (f *FileService) bootstrapNodes() error {
	for _, addr := range f.BootstrapNodes {
		log.Println("attempting to connect with node on", addr)
		go func(addr string) {
			if err := f.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err) 
			}
		}(addr)

	}

	return nil
}

func (f *FileService) listenForMessages() {
	defer func() {
		log.Println("file server stopped after user quit action")
		f.Transport.Close()
	}()

	for{
		select {
		case msg := <- f.Transport.Consume():
			fmt.Println(msg)
		case <- f.quitch:
			return
		}
	}
}