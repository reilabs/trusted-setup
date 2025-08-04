package phase2

import (
	"log"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
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

	log.Printf("Storing Phase 2 to %s", phase2Path)
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

	log.Printf("Loading Phase 2 from %s", phase2Path)
	_, err = phase2.ReadFrom(reader)
	if err != nil {
		return mpcsetup.Phase2{}, err
	}

	return
}

// SrsCommonsToFile writes circuit-independent components of the Groth16 SRS provided by srsCommons to a file
// specified by srsCommonsPath.
//
// Returns nil on success and error on failure.
func SrsCommonsToFile(srsCommons mpcsetup.SrsCommons, srsCommonsPath string) error {
	writer, err := os.Create(srsCommonsPath)
	if err != nil {
		return err
	}
	defer func(writer *os.File) {
		err := writer.Close()
		if err != nil {
			log.Printf("Error closing srsCommons writer: %v", err)
		}
	}(writer)

	log.Printf("Storing SRS commons to %s", srsCommonsPath)
	_, err = srsCommons.WriteTo(writer)
	if err != nil {
		return err
	}

	return nil
}

// SrsCommonsFromFile reads the circuit-independent components of the Groth16 SRS from the file specified by
// srsCommonsPath and returns the SrsCommons object.
//
// Returns the SrsCommons object and nil on success and nil and error on failure.
func SrsCommonsFromFile(srsCommonsPath string) (srsCommons mpcsetup.SrsCommons, err error) {
	reader, err := os.Open(srsCommonsPath)
	if err != nil {
		return mpcsetup.SrsCommons{}, err
	}
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {
			log.Printf("Error closing srsCommons reader: %v", err)
		}
	}(reader)

	log.Printf("Loading SRS commons from %s", srsCommonsPath)
	_, err = srsCommons.ReadFrom(reader)
	if err != nil {
		return mpcsetup.SrsCommons{}, err
	}

	return
}

// PkVkToFile writes the proving and verification keys to the files specified by pkFilePath and vkFilePath.
//
// Returns nil on success and error on failure.
func PkVkToFile(
	pk groth16.ProvingKey, pkFilePath string, vk groth16.VerifyingKey, vkFilePath string,
) error {
	pkWriter, err := os.Create(pkFilePath)
	if err != nil {
		return err
	}
	defer func(pkWriter *os.File) {
		err := pkWriter.Close()
		if err != nil {
			log.Printf("Error closing proving key writer: %v", err)
		}
	}(pkWriter)

	log.Printf("Storing Proving Key to %s", pkFilePath)
	_, err = pk.WriteTo(pkWriter)
	if err != nil {
		return err
	}

	vkWriter, err := os.Create(vkFilePath)
	if err != nil {
		return err
	}
	defer func(vkWriter *os.File) {
		err := vkWriter.Close()
		if err != nil {
			log.Printf("Error closing verifying key writer: %v", err)
		}
	}(vkWriter)

	log.Printf("Storing Verifying Key to %s", vkFilePath)
	_, err = vk.WriteTo(vkWriter)
	if err != nil {
		return err
	}

	return nil
}

// PkVkFromFile reads the proving and verification keys from the files specified by pkFilePath and vkFilePath.
//
// Returns the proving and verification keys and nil on success and empty keys and error on failure.
func PkVkFromFile(pkFilePath string, vkFilePath string) (groth16.ProvingKey, groth16.VerifyingKey, error) {
	pkReader, err := os.Open(pkFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer func(pkReader *os.File) {
		err := pkReader.Close()
		if err != nil {
			log.Printf("Error closing proving key reader: %v", err)
		}
	}(pkReader)

	log.Printf("Loading Proving Key from %s", pkFilePath)
	pk := groth16.NewProvingKey(ecc.BN254)
	_, err = pk.ReadFrom(pkReader)
	if err != nil {
		return nil, nil, err
	}

	vkReader, err := os.Open(vkFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer func(vkReader *os.File) {
		err := vkReader.Close()
		if err != nil {
			log.Printf("Error closing verifying key reader: %v", err)
		}
	}(vkReader)

	log.Printf("Loading Verifying Key from %s", vkFilePath)
	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err = vk.ReadFrom(vkReader)
	if err != nil {
		return nil, nil, err
	}

	return pk, vk, err
}
