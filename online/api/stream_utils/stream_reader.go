package stream_utils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/reilabs/trusted-setup/online/api"
)

type dataChunkReceiver interface {
	Recv() (*api.DataChunk, error)
}

type dataChunkResponseReceiver interface {
	Recv() (*api.ContributeResponse, error)
}

// NewStreamReader creates a new streamReader instance.
//
// The stream is meant to be a gRPC stream of either *api.DataChunk or *api.ContributeResponse.
func NewStreamReader(stream interface{}) io.Reader {
	return &streamReader{stream, bytes.Buffer{}}
}

type streamReader struct {
	stream interface{}
	buffer bytes.Buffer
}

// Read implements the io.Reader interface. It reads data from the stream set by NewStreamReader into p.
func (d *streamReader) Read(p []byte) (n int, err error) {
	switch d.stream.(type) {
	case dataChunkReceiver:
		if d.buffer.Len() == 0 {
			resp, err := d.stream.(dataChunkReceiver).Recv()
			if err != nil {
				return 0, err
			}
			d.buffer.Write(resp.Data)
		}
	case dataChunkResponseReceiver:
		if d.buffer.Len() == 0 {
			resp, err := d.stream.(dataChunkResponseReceiver).Recv()
			if err != nil {
				return 0, err
			}
			d.buffer.Write(resp.GetLastContribution().Data)
		}
	default:
		return 0, fmt.Errorf("unsupported stream type: %T", d.stream)
	}

	return d.buffer.Read(p)
}
