// Package storage provides a persistent storage service with multiple backends.
//
// Currently supported backends are Tmpfs and AWS S3.
//
// Objects stored in the storage must implement io.WriterTo interface.
package storage

import "io"

type Storage interface {
	Save(id string, obj io.WriterTo) (string, error)
	List() ([]string, error)
}
