package main

import (
	"bytes"
	"fmt"
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

func TestStore(t *testing.T) {
	s:= newStore()
	defer teardown(t, s)

	data := []byte("some file bytes")
	for i:= 0; i< 10; i++{
		key := fmt.Sprintf("MyFileKey_%d", i)
	
		err := s.writeStream(key, bytes.NewBuffer(data))
		assert.NoError(t, err, "expect writeStream to run without errors")

		hasIt := s.Has(key)
		assert.True(t, hasIt, "expect file storage to has %s key", key)
	
		r, err := s.Read(key)
		assert.NoError(t, err, "expect Read to run without errors")
	
		b, _ := io.ReadAll(r)
		assert.Equal(t, b, data, "expect the file content to be the same as the provided one")

		err = s.Delete(key)
		assert.NoError(t, err, "expect Delete to run without errors")
	}

}

func newStore() *Store {
	opts := StoreOpts{
		PathTrasnformFunc: CASPathTransformFunc,
	}

	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
