package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
)

func Phase2Verify(_ context.Context, cmd *cli.Command) error {
	phase2FilePaths := cmd.StringSlice("phase2")

	return phase2.Verify(phase2FilePaths)
}
