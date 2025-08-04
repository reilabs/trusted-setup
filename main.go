package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/offline"
)

func main() {
	app := &cli.Command{
		Name:  filepath.Base(os.Args[0]),
		Usage: "a ZKP Trusted Setup Ceremony Coordinator",
		Description: "This program allows for initializing a trusted setup ceremony and contributing to it.\n" +
			"Phase 2 of the ceremony can be initialized from a previously generated Phase 1 file\n" +
			"or from a Snarkjs powers of tau file. New contributions can be added to Phase 2.\n" +
			"The contributions can be verified. Proving and verifying keys can be exported from the\n" +
			"ceremony artifacts.\n\n" +
			"Note that, as for now, the program requires the input constraint system to be produced\n" +
			"by Gnark v0.13. The used backend must be Groth16 and the elliptic curve used must be BN254.",
		Authors: []any{
			"Wojciech Żmuda <wojciech.zmuda@reilabs.io>",
		},
		Copyright: "(c) 2025 Reilabs sp. z o.o.",
		Action: func(context.Context, *cli.Command) error {
			return nil
		},
		DefaultCommand: "help",
		Suggest:        true,
		Commands:       offline.Commands,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
	log.Print("Operation successful")
}
