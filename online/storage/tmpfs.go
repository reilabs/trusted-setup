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
	defer func(tmpFile *os.File) {
		err = tmpFile.Close()
		if err != nil {
			return
		}
	}(tmpFile)

	_, err = obj.WriteTo(tmpFile)
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	t.files = append(t.files, tmpFile.Name())

	return tmpFile.Name(), nil
}

// List returns a list of all files stored with Save.
//
// The function reflects the actual state of the storage. I.e. if the file stored with Save
// is removed with `rm /tmp/<file name>`, List will not return it.
//
// Returns an array of strings where each element is the id of the file.
// Error is always nil.
func (t *Tmpfs) List() ([]string, error) {
	var existingFiles []string
	for _, f := range t.files {
		fi, err := os.Stat(f)
		if err == nil && !fi.IsDir() {
			existingFiles = append(existingFiles, f)
		}
	}

	return existingFiles, nil
}
