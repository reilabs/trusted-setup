// Package phase1 implements functions interacting with multi-party computation Phase 1 objects produced by Gnark.
package phase1

import (
	"log"

	deserializer "github.com/worldcoin/ptau-deserializer/deserialize"
)

// FromPtau reads the Starkjs powers of tau object from the file specified by ptauFilePath, converts it to a Phase 1
// object and writes the Phase 1 object to the file specified by outputPhase1FilePath.
//
// Returns nil on success and error on failure.
func FromPtau(ptauFilePath string, outputPhase1FilePath string) error {
	log.Printf("Loading Starkjs powers of tau from %s", ptauFilePath)
	ptau, err := deserializer.ReadPtau(ptauFilePath)
	if err != nil {
		return err
	}

	log.Print("Converting Starkjs powers of tau to Phase 1")
	phase1, err := deserializer.ConvertPtauToPhase1(ptau)
	if err != nil {
		return err
	}

	return ToFile(&phase1, outputPhase1FilePath)
}
