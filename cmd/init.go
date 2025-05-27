package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
)

func Phase2Init(_ context.Context, cmd *cli.Command) error {
	phase1FilePath := cmd.String("phase1")
	r1csFilePath := cmd.String("r1cs")
	outputPhase2FilePath := cmd.String("phase2")
	outputEvalFilePath := cmd.String("eval")

	return phase2.Init(phase1FilePath, r1csFilePath, outputPhase2FilePath, outputEvalFilePath)
}
