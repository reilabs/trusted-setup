package storage_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/storage"
)

func TestTmpfs_Save(t *testing.T) {
	tmpfs := storage.NewTmpfs("")

	content := []byte("hello world")
	obj := bytes.NewBuffer(content)

	path, err := tmpfs.Save("testfile", obj)
	defer os.Remove(path)
	assert.NoError(t, err)

	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.False(t, info.IsDir(), "Expected a file but found a directory")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	assert.True(t, bytes.Equal(data, content), "File content mismatch")
}

func TestTmpfs_List(t *testing.T) {
	tmpfs := storage.NewTmpfs("")

	content := []byte("hello world")
	obj := bytes.NewBuffer(content)

	assert.Empty(t, tmpfs.List())

	for i := 0; i < 5; i++ {
		path, err := tmpfs.Save(fmt.Sprintf("testfile%d", i), obj)
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
		assert.Len(t, tmpfs.List(), i+1)
	}

	for _, f := range tmpfs.List() {
		os.Remove(f)
	}
}
