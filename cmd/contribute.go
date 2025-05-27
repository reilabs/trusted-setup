package cmd

import (
	"context"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
)

func Phase2Contribute(_ context.Context, cmd *cli.Command) error {
	phase2FilePath := cmd.String("phase2")

	newFileName, err := phase2.Contribute(phase2FilePath)
	if err != nil {
		return err
	}

	log.Printf("Phase2 file with contributions: %s", newFileName)
	return nil
}
