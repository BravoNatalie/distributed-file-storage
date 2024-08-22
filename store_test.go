package main

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T){
	key := "ImportantFile"
	filePath := CASPathTransformFunc(key)

	expectedDirectory := "2e994/28bf7/75b5b/af283/851ec/04d27/b709f/ae691"
	expectedFilename := "2e99428bf775b5baf283851ec04d27b709fae691"

	assert.Equal(t, filePath.Directory, expectedDirectory)
	assert.Equal(t, filePath.Filename, expectedFilename)
}

func TestDeleteKey(t *testing.T){
	opts := StoreOpts{
		PathTrasnformFunc: CASPathTransformFunc,
	}

	s:= NewStore(opts)

	key := "MyFileKey"
	data := []byte("some file bytes")

	assert.NoError(t, s.writeStream(key, bytes.NewBuffer(data)))
	assert.NoError(t, s.Delete(key))
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTrasnformFunc: CASPathTransformFunc,
	}

	s:= NewStore(opts)

	key := "MyFileKey"
	data := []byte("some file bytes")

	assert.NoError(t, s.writeStream(key, bytes.NewBuffer(data)))

	r, err := s.Read(key)
	assert.NoError(t, err)

	hasIt := s.Has(key)
	assert.True(t, hasIt)

	b, _ := io.ReadAll(r)
	assert.Equal(t, b, data)

	s.Delete(key)
}