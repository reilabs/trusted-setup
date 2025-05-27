// Package r1cs implements methods interacting with the R1CS constraint system object produced by Gnark.
package r1cs

import (
	"log"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
)

// ToFile writes the R1CS constraint system provided in ccs to a file specified by r1csPath.
//
// Returns nil on success and error on failure.
func ToFile(ccs constraint.ConstraintSystem, r1csPath string) error {
	writer, err := os.Create(r1csPath)
	if err != nil {
		return err
	}
	defer func(writer *os.File) {
		err := writer.Close()
		if err != nil {
			log.Printf("Error closing r1cs writer: %v", err)
		}
	}(writer)

	_, err = ccs.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

// FromFile reads the R1CS constraint system from the file specified by r1csPath and returns the constraint system
// object.
//
// Returns the constraint system object and nil on success and nil and error on failure.
func FromFile(r1csPath string) (constraint.ConstraintSystem, error) {
	reader, err := os.Open(r1csPath)
	if err != nil {
		return nil, err
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			log.Printf("Error closing r1cs reader: %v", err)
		}
	}(reader)

	r1cs := groth16.NewCS(ecc.BN254)
	_, err = r1cs.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	return r1cs, nil
}
