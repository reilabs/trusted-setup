package main

import (
	"log"
	"os"
	"path/filepath"

	_ "github.com/consensys/gnark"
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

func phase1FromPtau(ptau string, r1cs string, phase2 string) error {
	// Note: The current implementation does not use r1cs
	file, err := deserializer.InitPtau(ptau)
	if err != nil {
		return err
	}
	err = deserializer.WritePhase1FromPtauFile(file, phase2)
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
