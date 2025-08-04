package stream_utils

import (
	"fmt"
	"io"

	"github.com/reilabs/trusted-setup/online/api"
)

type dataChunkSender interface {
	Send(contribution *api.DataChunk) error
}

type dataChunkResponseSender interface {
	Send(contribution *api.ContributeResponse) error
}

// NewStreamWriter creates a new streamWriter instance.
//
// The stream is meant to be a gRPC stream of either *api.DataChunk or *api.ContributeResponse.
func NewStreamWriter(stream interface{}) io.Writer {
	return &streamWriter{stream}
}

type streamWriter struct {
	stream interface{}
}

// Write implements the io.Writer interface. It sends the data from p to the stream set by NewStreamWriter.
func (u *streamWriter) Write(p []byte) (n int, err error) {
	switch u.stream.(type) {
	case dataChunkSender:
		err = u.stream.(dataChunkSender).Send(api.NewDataChunk(p))
		if err != nil {
			return 0, err
		}
	case dataChunkResponseSender:
		err = u.stream.(dataChunkResponseSender).Send(api.NewLastContribution(p))
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("unsupported stream type: %T", u.stream)
	}

	return len(p), nil
}
