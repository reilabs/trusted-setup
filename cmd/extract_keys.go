package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/keys"
)

func ExtractKeys(_ context.Context, cmd *cli.Command) error {
	phase1FilePath := cmd.String("phase1")
	phase2FilePath := cmd.String("phase2")
	evalFilePath := cmd.String("eval")
	r1csFilePath := cmd.String("r1cs")
	pkFilePath := cmd.String("pk")
	vkFilePath := cmd.String("vk")

	return keys.Extract(phase1FilePath, phase2FilePath, evalFilePath, r1csFilePath, pkFilePath, vkFilePath)
}
