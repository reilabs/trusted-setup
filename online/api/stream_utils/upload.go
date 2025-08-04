package stream_utils

import (
	"github.com/reilabs/trusted-setup/online/api"
)

type dataChunkSender interface {
	Send(*api.DataChunk) error
}

type StreamUploader interface {
	Write(p []byte) (n int, err error)
}

func NewStreamUploader(stream dataChunkSender) StreamUploader {
	return streamUploader{stream}
}

type streamUploader struct {
	Stream dataChunkSender
}

func (u streamUploader) Write(p []byte) (n int, err error) {
	err = u.Stream.Send(&api.DataChunk{Data: p})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
