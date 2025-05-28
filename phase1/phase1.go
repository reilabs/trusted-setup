package phase1

import (
	"log"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bn254"
	deserializer "github.com/worldcoin/ptau-deserializer/deserialize"
)

func phase2FromPtau(ptau deserializer.Ptau, r1cs constraint.ConstraintSystem) (mpcsetup.Phase2, error) {
	phase1, err := deserializer.ConvertPtauToPhase1(ptau)
	if err != nil {
		return mpcsetup.Phase2{}, err
	}

	phase2, _ := mpcsetup.InitPhase2(r1cs.(*cs.R1CS), &phase1)

	return phase2, nil
}

func r1csFromPath(r1csPath string) (constraint.ConstraintSystem, error) {
	r1csReader, err := os.Open(r1csPath)
	if err != nil {
		return nil, err
	}
	defer func(r1csReader *os.File) {
		err := r1csReader.Close()
		if err != nil {
			log.Printf("Error closing r1cs reader: %v", err)
		}
	}(r1csReader)

	r1cs := groth16.NewCS(ecc.BN254)
	_, err = r1cs.ReadFrom(r1csReader)
	if err != nil {
		return nil, err
	}

	return r1cs, nil
}

func phase1FromPath(phase1Path string) (mpcsetup.Phase1, error) {
	phase1Reader, err := os.Open(phase1Path)
	if err != nil {
		return mpcsetup.Phase1{}, err
	}
	defer func(path1Reader *os.File) {
		err := path1Reader.Close()
		if err != nil {
			log.Printf("Error closing phase1 reader: %v", err)
		}
	}(phase1Reader)

	var phase1 mpcsetup.Phase1
	_, err = phase1.ReadFrom(phase1Reader)
	if err != nil {
		return mpcsetup.Phase1{}, err
	}

	return phase1, nil
}

func phase2toFile(phase2 mpcsetup.Phase2, phase2Path string) error {
	phase2Writer, err := os.Create(phase2Path)
	if err != nil {
		return err
	}
	defer func(phase2Writer *os.File) {
		err := phase2Writer.Close()
		if err != nil {
			log.Printf("Error closing phase2 writer: %v", err)
		}
	}(phase2Writer)
	_, err = phase2.WriteTo(phase2Writer)
	if err != nil {
		return err
	}

	return nil
}

func FromPtauPath(ptauPath string, r1csPath string, phase2Path string) error {
	ptau, err := deserializer.ReadPtau(ptauPath)
	if err != nil {
		return err
	}

	r1cs, err := r1csFromPath(r1csPath)
	if err != nil {
		return err
	}

	phase2, err := phase2FromPtau(ptau, r1cs)
	if err != nil {
		return err
	}

	return phase2toFile(phase2, phase2Path)
}

func FromPhase1Path(phase1Path string, r1csPath string, phase2Path string) error {
	r1cs, err := r1csFromPath(r1csPath)
	if err != nil {
		return err
	}

	phase1, err := phase1FromPath(phase1Path)
	if err != nil {
		return err
	}

	phase2, _ := mpcsetup.InitPhase2(r1cs.(*cs.R1CS), &phase1)

	return phase2toFile(phase2, phase2Path)
}
