package storage

import "io"

type Storage interface {
	Save(id string, obj io.WriterTo) (string, error)
	List() []string
}
