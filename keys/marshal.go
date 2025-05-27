package keys

import (
	"log"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
)

// ToFile writes the proving and verification keys to the files specified by pkFilePath and vkFilePath.
//
// Returns nil on success and error on failure.
func ToFile(
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
	_, err = vk.WriteTo(vkWriter)
	if err != nil {
		return err
	}

	return nil
}

// FromFile reads the proving and verification keys from the files specified by pkFilePath and vkFilePath.
//
// Returns the proving and verification keys and nil on success and empty keys and error on failure.
func FromFile(pkFilePath string, vkFilePath string) (groth16.ProvingKey, groth16.VerifyingKey, error) {
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

	vk := groth16.NewVerifyingKey(ecc.BN254)
	_, err = vk.ReadFrom(vkReader)
	if err != nil {
		return nil, nil, err
	}

	return pk, vk, err
}
