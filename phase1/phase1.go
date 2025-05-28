package phase1

import (
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"
	deserializer "github.com/worldcoin/ptau-deserializer/deserialize"
)

func FromPtauPath(ptauPath string, r1csPath string, phase2Path string) error {
	ptau, err := deserializer.ReadPtau(ptauPath)
	if err != nil {
		return err
	}

	phase1, err := deserializer.ConvertPtauToPhase1(ptau)
	if err != nil {
		return err
	}

	r1csReader, err := os.Open(r1csPath)
	if err != nil {
		return err
	}
	defer func(r1csReader *os.File) {
		err := r1csReader.Close()
		if err != nil {

		}
	}(r1csReader)
	r1cs := groth16.NewCS(ecc.BN254)
	_, err = r1cs.ReadFrom(r1csReader)
	if err != nil {
		return err
	}

	phase2, _ := mpcsetup.InitPhase2(r1cs.(*cs.R1CS), &phase1)
	phase2File, err := os.Create(phase2Path)
	if err != nil {
		return err
	}
	_, err = phase2.WriteTo(phase2File)
	if err != nil {
		return err
	}

	return nil
}

func FromPhase1Path(phase1 string, r1cs string, phase2 string) error {
	// This function needs to be implemented
	// Here, further processing should be done using the provided r1cs and phase1 files
	return nil
}
