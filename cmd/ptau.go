package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase1"
)

func PtauToPhase1(_ context.Context, cmd *cli.Command) error {
	ptauFilePath := cmd.String("ptau")
	outputPhase1FilePath := cmd.String("phase1")
	log.Printf(
		"Convert Starkjs powers of tau to Phase 1:\n"+
			"\tLoad ptau from:   %s\n"+
			"\tStore Phase 1 to: %s\n",
		ptauFilePath,
		outputPhase1FilePath,
	)
	if ptauFilePath == "" || outputPhase1FilePath == "" {
		return fmt.Errorf("one of the required file paths is empty")
	}

	return phase1.FromPtau(ptauFilePath, outputPhase1FilePath)
}
