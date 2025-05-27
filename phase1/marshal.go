package phase1

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"os"

	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
)

// ToFile writes the Phase 1 object to the file specified by phase1Path.
//
// Returns nil on success and error on failure.
func ToFile(phase1 mpcsetup.Phase1, phase1Path string) error {
	// Phase1 may not have the hash field populated (probably a bug in ptau-deserializer).
	// Populate it if it's empty before serializing.
	if len(phase1.Hash) == 0 {
		var serializedPhase1 bytes.Buffer
		enc := gob.NewEncoder(&serializedPhase1)
		err := enc.Encode(phase1.Parameters)
		if err != nil {
			return err
		}
		err = enc.Encode(phase1.PublicKeys)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(serializedPhase1.Bytes())
		phase1.Hash = hash[:]
	}

	writer, err := os.Create(phase1Path)
	if err != nil {
		return err
	}
	defer func(writer *os.File) {
		err := writer.Close()
		if err != nil {
			log.Printf("Error closing phase1 writer: %v", err)
		}
	}(writer)
	_, err = phase1.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

// FromFile reads the Phase 1 object from the file specified by phase1Path.
//
// Returns the Phase 1 object and nil on success and empty Phase 1 object and error on failure.
func FromFile(phase1Path string) (phase1 mpcsetup.Phase1, err error) {
	reader, err := os.Open(phase1Path)
	if err != nil {
		return mpcsetup.Phase1{}, err
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			log.Printf("Error closing phase1 reader: %v", err)
		}
	}(reader)

	_, err = phase1.ReadFrom(reader)
	if err != nil {
		return mpcsetup.Phase1{}, err
	}

	return
}
