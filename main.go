package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/reilabs/whir-trusted-setup/phase2"
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
						return phase2.FromPtauPath(ptauFilePath, r1csFilePath, outputPhase2FilePath)
					} else if phase1FilePath != "" && ptauFilePath == "" {
						return phase2.FromPhase1Path(phase1FilePath, r1csFilePath, outputPhase2FilePath)
					}

					log.Fatal("Invalid input: either ptau or phase1 file must be specified, but not both.")
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "ptau",
						Usage: "Snarkjs powers of tau file (required if not using --phase1 flag)",
					},
					&cli.StringFlag{
						Name:  "phase1",
						Usage: "Phase 1 file (required if not using --ptau flag)",
					},
					&cli.StringFlag{
						Name:     "r1cs",
						Usage:    "R1CS file generated from gnark-whir",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "phase2",
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
