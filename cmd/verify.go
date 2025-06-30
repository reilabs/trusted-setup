package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/phase2"
)

func Phase2Verify(_ context.Context, cmd *cli.Command) error {
	phase2prevFilePath := cmd.String("phase2prev")
	phase2nextFilePath := cmd.String("phase2next")
	log.Printf(
		"Verify single Phase 2 contribution:\n"+
			"\tLoad previous Phase 2 from: %s\n"+
			"\tLoad next Phase 2 from:     %s\n",
		phase2prevFilePath,
		phase2nextFilePath,
	)
	if phase2prevFilePath == "" || phase2nextFilePath == "" {
		return fmt.Errorf("one of the required file paths is empty")
	}

	return phase2.Verify(phase2prevFilePath, phase2nextFilePath)
}

func Phase2ExtractKeys(_ context.Context, cmd *cli.Command) error {
	r1csFilePath := cmd.String("r1cs")
	srsCommonsFilePath := cmd.String("srscommons")
	phase2FilePaths := cmd.StringSlice("phase2")
	pkFilePath := cmd.String("pk")
	vkFilePath := cmd.String("vk")
	log.Printf(
		"Verify multiple Phase 2 contributions:\n"+
			"\tLoad R1CS from:         %s\n"+
			"\tLoad SRS commons from:  %s\n"+
			"\tLoad Phase 2 from:      %s\n"+
			"\tStore Proving Key to:   %s\n"+
			"\tStore Verifying Key to: %s\n",
		r1csFilePath,
		srsCommonsFilePath,
		phase2FilePaths,
		pkFilePath,
		vkFilePath,
	)
	if r1csFilePath == "" || srsCommonsFilePath == "" || len(phase2FilePaths) == 0 || pkFilePath == "" || vkFilePath == "" {
		return fmt.Errorf("one of the required file paths is empty")
	}

	return phase2.ExtractKeys(r1csFilePath, srsCommonsFilePath, phase2FilePaths, pkFilePath, vkFilePath)
}
