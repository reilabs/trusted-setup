// Package storage provides a persistent storage service with multiple backends.
//
// Objects stored in the storage must implement io.WriterTo interface.
package storage

import "io"

type Storage interface {
	Save(id string, obj io.WriterTo) (string, error)
	List() ([]string, error)
}
