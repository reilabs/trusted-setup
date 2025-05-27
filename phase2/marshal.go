package phase2

import (
	"io"
	"log"
	"os"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
)

// ToFile writes the Phase 2 object provided in phase2 to a file specified by phase2Path.
//
// Returns nil on success and error on failure.
func ToFile(phase2 mpcsetup.Phase2, phase2Path string) error {
	writer, err := os.Create(phase2Path)
	if err != nil {
		return err
	}
	defer func(writer *os.File) {
		err := writer.Close()
		if err != nil {
			log.Printf("Error closing phase2 writer: %v", err)
		}
	}(writer)
	_, err = phase2.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

// FromFile reads the Phase 2 object from the file specified by phase2Path and returns the Phase 2 object.
//
// Returns the Phase 2 object and nil on success and nil and error on failure.
func FromFile(phase2Path string) (phase2 mpcsetup.Phase2, err error) {
	reader, err := os.Open(phase2Path)
	if err != nil {
		return mpcsetup.Phase2{}, err
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			log.Printf("Error closing phase2 reader: %v", err)
		}
	}(reader)

	_, err = phase2.ReadFrom(reader)
	if err != nil {
		return mpcsetup.Phase2{}, err
	}

	return
}

// EvalToFile writes the Phase 2 evaluations object provided in eval to a file specified by evalPath.
//
// Returns nil on success and error on failure.
func EvalToFile(eval mpcsetup.Phase2Evaluations, evalPath string) error {
	writer, err := os.Create(evalPath)
	if err != nil {
		return err
	}
	defer func(writer *os.File) {
		err := writer.Close()
		if err != nil {
			log.Printf("Error closing phase2 evaluations writer: %v", err)
		}
	}(writer)
	_, err = eval.WriteTo(writer)
	if err != nil {
		return err
	}

	// Gnark has a bug - it does not serialize VKK from eval's G1.
	// Due to this the number of public variables in Verify() is miscalculated,
	// and proof verification fails.
	// TODO: remove this workaround when we update to Gnark v0.13.0.
	for _, g1affine := range eval.G1.VKK {
		buffer := g1affine.RawBytes()
		_, err = writer.Write(buffer[:])
		if err != nil {
			return err
		}
	}

	return nil
}

// EvalFromFile reads the Phase 2 object from the file specified by evalPath and returns the Phase 2 evaluations object.
//
// Returns the Phase 2 evaluations object and nil on success and nil and error on failure.
func EvalFromFile(evalPath string) (eval mpcsetup.Phase2Evaluations, err error) {
	reader, err := os.Open(evalPath)
	if err != nil {
		return mpcsetup.Phase2Evaluations{}, err
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			log.Printf("Error closing phase2 evaluations reader: %v", err)
		}
	}(reader)

	_, err = eval.ReadFrom(reader)
	if err != nil {
		return mpcsetup.Phase2Evaluations{}, err
	}

	// Gnark has a bug - it does not serialize VKK from eval's G1.
	// Due to this the number of public variables in Verify() is miscalculated,
	// and proof verification fails.
	// TODO: remove this workaround when we update to Gnark v0.13.0.
	for {
		buffer := make([]byte, 64)
		_, err = io.ReadFull(reader, buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return mpcsetup.Phase2Evaluations{}, err
		}
		var g1 bn254.G1Affine
		_, err = g1.SetBytes(buffer)
		if err != nil {
			return mpcsetup.Phase2Evaluations{}, err
		}
		eval.G1.VKK = append(eval.G1.VKK, g1)
	}

	return eval, nil
}
