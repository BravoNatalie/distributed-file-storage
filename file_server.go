package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
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

func (f *FileService) StoreData(key string, r io.Reader) error {

	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	if err := f.store.Write(key, tee); err != nil {
		return err
	}

	p := &Payload{
		Key: key,
		Data: buf.Bytes(),
	}

	fmt.Println(buf.Bytes())

	return f.broadcast(p)
}

type Payload struct {
	Key string
	Data []byte
}

func (f *FileService) broadcast(p *Payload) error {
	peers := []io.Writer{}
	for _, peer := range f.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return  gob.NewEncoder(mw).Encode(p) 
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
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil{
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", p)
		case <- f.quitch:
			return
		}
	}
}