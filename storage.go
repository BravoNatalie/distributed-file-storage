package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName =  "dfs_root"

func CASPathTransformFunc(key string) FilePath {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:]) // convert array to slice

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+ blockSize
		paths[i] = hashStr[from:to]
	}

	return FilePath {
		Directory: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTrasnformFunc func(string) FilePath

type FilePath struct {
	Directory string
	Filename string
}

func (p *FilePath) FullPath() string{
	return fmt.Sprintf("%s/%s", p.Directory, p.Filename)
}

type StoreOpts struct {
	// StorageRoot is the parent folder name, containing all the folders/files of the system
	StorageRoot string
	PathTrasnformFunc PathTrasnformFunc
}

var DefaultPathTransformFunc = func(key string) FilePath {
	return FilePath {
		Directory: key,
		Filename: key,
	}
}

type Store struct{
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTrasnformFunc == nil {
		opts.PathTrasnformFunc = DefaultPathTransformFunc
	}
 
	if opts.StorageRoot == "" {
		opts.StorageRoot = defaultRootFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.StorageRoot)
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key,r)
}

func (s *Store) Has(key string) bool {
	filePath := s.PathTrasnformFunc(key) 

	fullPath := fmt.Sprintf("%s/%s", s.StorageRoot, filePath.FullPath())

	_, err := os.Stat(fullPath)
	return err != fs.ErrNotExist
}

func (s *Store) Delete(key string) error {
	filePath := s.PathTrasnformFunc(key) 

	parentFolder := s.StorageRoot + "/" + strings.Split(filePath.Directory, "/")[0] + "/"

	defer func() {
		log.Printf("Successfully deleted key [%s] in directory [%s] from disk.", filePath.Filename, parentFolder)
	}()

	return os.RemoveAll(parentFolder)
}

func (s *Store) Read(key string) (io.Reader, error){
	f, err := s.getFile(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) getFile(key string) (*os.File, error) {
	filePath := s.PathTrasnformFunc(key)
	fullPath := fmt.Sprintf("%s/%s", s.StorageRoot, filePath.FullPath())
	return os.Open(fullPath)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	filePath := s.PathTrasnformFunc(key)

	fullDirectory := fmt.Sprintf("%s/%s", s.StorageRoot, filePath.Directory)

	if err := os.MkdirAll(fullDirectory, os.ModePerm); err != nil {
		return err
	}

	fullPath := fmt.Sprintf("%s/%s", s.StorageRoot, filePath.FullPath())

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f,r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, fullPath)

	return nil
}
