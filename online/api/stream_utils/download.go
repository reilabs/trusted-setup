package stream_utils

import (
	"bytes"

	"github.com/reilabs/trusted-setup/online/api"
)

type dataChunkReceiver interface {
	Recv() (*api.DataChunk, error)
}

type StreamDownloader interface {
	Read(p []byte) (n int, err error)
}

func NewStreamDownloader(stream dataChunkReceiver) StreamDownloader {
	return streamDownloader{stream, bytes.Buffer{}}
}

type streamDownloader struct {
	Stream dataChunkReceiver
	buffer bytes.Buffer
}

func (sr streamDownloader) Read(p []byte) (n int, err error) {
	if sr.buffer.Len() == 0 {
		chunk, err := sr.Stream.Recv()
		if err != nil {
			return 0, err
		}

		sr.buffer.Write(chunk.Data)
	}

	n, err = sr.buffer.Read(p)
	return n, err
}
