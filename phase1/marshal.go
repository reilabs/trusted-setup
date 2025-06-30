package phase1

import (
	"log"
	"os"

	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
)

// ToFile writes the Phase 1 object to the file specified by phase1Path.
//
// Returns nil on success and error on failure.
func ToFile(phase1 mpcsetup.Phase1, phase1Path string) error {
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

	log.Printf("Storing Phase 1 to %s", phase1Path)
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

	log.Printf("Loading Phase 1 from %s", phase1Path)
	_, err = phase1.ReadFrom(reader)
	if err != nil {
		return mpcsetup.Phase1{}, err
	}

	return
}
