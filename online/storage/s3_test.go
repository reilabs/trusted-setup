package storage_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/storage"
)

type ExampleWriterTo struct {
	content []byte
}

func (e *ExampleWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(e.content)
	return int64(n), err
}

func TestS3(t *testing.T) {
	s := storage.NewS3("trusted-setup-ceremony-v2", "us-east-1")

	data := &ExampleWriterTo{content: []byte("Hello world")}
	filePath, err := s.Save("test", data)
	assert.NoError(t, err)
	assert.Equal(t, "s3://trusted-setup-ceremony-v2/test", filePath)

	files := s.List()
	assert.Len(t, files, 1)
	assert.Equal(t, files[0], filePath)
}
