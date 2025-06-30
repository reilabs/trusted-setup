package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
)

func Phase2Contribute(_ context.Context, cmd *cli.Command) error {
	phase2FilePath := cmd.String("phase2")
	log.Printf(
		"Contribution to Phase 2:\n"+
			"\tLoad Phase 2 from:    %s\n",
		phase2FilePath,
	)
	if phase2FilePath == "" {
		return fmt.Errorf("input Phase 2 file path is empty")
	}
	newFileName, err := phase2.Contribute(phase2FilePath)
	if err != nil {
		return err
	}

	log.Printf("Phase2 file with contributions: %s", newFileName)
	return nil
}
