package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
	"github.com/reilabs/trusted-setup/utils/randomness"
)

func Phase2Init(_ context.Context, cmd *cli.Command) error {
	rand, err := randomness.NewDrandProvider()
	if err != nil {
		return err
	}
	beacon := rand.GetBeacon()

	phase1FilePath := cmd.String("phase1")
	r1csFilePath := cmd.String("r1cs")
	outputPhase2FilePath := cmd.String("phase2")
	outputSrsCommonsFilePath := cmd.String("srscommons")
	log.Printf(
		"Initializing Phase 2:\n"+
			"\tLoad Phase 1 from:                %s\n"+
			"\tLoad R1CS from:                   %s\n"+
			"\tStore Phase 2 to:                 %s\n"+
			"\tStore SRS commons to:             %s\n"+
			"\tBeacon (pass it to extract-keys): %x\n",
		phase1FilePath, r1csFilePath, outputPhase2FilePath, outputSrsCommonsFilePath, beacon,
	)
	if phase1FilePath == "" || r1csFilePath == "" || outputPhase2FilePath == "" || outputSrsCommonsFilePath == "" {
		return fmt.Errorf("one of the required file paths is empty")
	}

	return phase2.Init(
		phase1FilePath, r1csFilePath, outputPhase2FilePath, outputSrsCommonsFilePath, beacon,
	)
}
