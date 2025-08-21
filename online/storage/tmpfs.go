package storage

import (
	"io"
	"os"
)

type Tmpfs struct {
	prefix string
	files  []string
}

// NewTmpfs creates a new instance of the storage backed by tmpfs.
//
// Prefix is a prepended to the names of the files stored with Save.
func NewTmpfs(prefix string) *Tmpfs {
	return &Tmpfs{
		prefix,
		[]string{},
	}
}

// Save stores the object in the temp filesystem.
//
// The given id is prepended to the file name of the file. If the file already exists, it will
// be overwritten.
//
// The function returns the path to the file and an error if any.
func (t *Tmpfs) Save(id string, obj io.WriterTo) (string, error) {
	tmpFile, err := os.CreateTemp("", t.prefix+"-"+id+"-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = obj.WriteTo(tmpFile)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	t.files = append(t.files, tmpFile.Name())

	return tmpFile.Name(), nil
}

// List returns a list of all files stored with Save.
func (t *Tmpfs) List() []string {
	return t.files
}
