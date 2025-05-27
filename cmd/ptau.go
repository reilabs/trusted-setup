package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase1"
)

func PtauToPhase1(_ context.Context, cmd *cli.Command) error {
	ptauFilePath := cmd.String("ptau")
	outputPhase1FilePath := cmd.String("phase1")

	return phase1.FromPtau(ptauFilePath, outputPhase1FilePath)
}
