package main

import (
	"log"
	"os"
	"path/filepath"

	_ "github.com/consensys/gnark"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/urfave/cli/v2"
	_ "github.com/worldcoin/ptau-deserializer/deserialize"
	deserializer "github.com/worldcoin/ptau-deserializer/deserialize"
)

func main() {
	app := &cli.App{
		Name:  filepath.Base(os.Args[0]),
		Usage: "Trusted setup ceremony coordinator for github.com/reilabs/gnark-whir",
		Action: func(*cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "init",
				Usage: "Initialize phase 2 for the given R1CS from gnark-whir circuit with either phase 1 or powers" +
					" of tau file",
				Action: func(cCtx *cli.Context) error {
					phase1FilePath := cCtx.String("phase1")
					ptauFilePath := cCtx.String("ptau")
					r1csFilePath := cCtx.String("r1cs")
					outputPhase2FilePath := cCtx.String("phase2")

					if ptauFilePath != "" && phase1FilePath == "" {
						return phase1FromPtau(ptauFilePath, r1csFilePath, outputPhase2FilePath)
					} else if phase1FilePath != "" && ptauFilePath == "" {
						return phase1FromPh1(phase1FilePath, r1csFilePath, outputPhase2FilePath)
					}

					log.Fatal("Invalid input: either ptau or phase1 file must be specified, but not both.")
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "ptau",
						Usage: "Load a snarkjs powers of tau file to convert it to a phase 1 file " +
							"(required if not using --phase1 flag)",
					},
					&cli.StringFlag{
						Name:  "phase1",
						Usage: "Phase 1 file path (required if not using --ptau flag)",
					},
					&cli.StringFlag{
						Name:     "r1cs",
						Usage:    "R1CS file path",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Output path for the phase 2 file",
						Required: true,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func phase1FromPtau(ptauPath string, r1csPath string, phase2Path string) error {
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

func phase1FromPh1(phase1 string, r1cs string, phase2 string) error {
	// This function needs to be implemented
	// Here, further processing should be done using the provided r1cs and phase1 files
	return nil
}
