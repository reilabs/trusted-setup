package offline

import (
	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/offline/actions"
)

// Commands defines all CLI commands related to offline operations.
var Commands = []*cli.Command{
	{
		Name:   "ptau",
		Usage:  "Convert a Snarkjs powers of tau file to a Phase 1 file",
		Action: actions.PtauToPhase1,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ptau",
				Usage:    "Snarkjs powers of tau file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "phase1",
				Usage:    "Output Phase 1 file",
				Required: true,
			},
		},
	},
	{
		Name:     "init",
		Category: "offline mode",
		Usage:    "Initialize Phase 2 for the given R1CS with a Phase 1 file",
		Action:   actions.Phase2Init,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "phase1",
				Usage:    "Phase 1 file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "r1cs",
				Usage:    "R1CS file generated from a Gnark circuit",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "phase2",
				Usage:    "Output path for the Phase 2 file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "srscommons",
				Usage:    "Output path for circuit-independent components of the Groth16 SRS",
				Required: true,
			},
		},
	},
	{
		Name:     "contribute",
		Category: "offline mode",
		Usage:    "Contribute randomness to Phase 2",
		Action:   actions.Phase2Contribute,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "phase2",
				Usage: "The existing Phase 2 file created in the init step or in the previous run\n" +
					"of the contribute step.",
				Required: true,
			},
		},
	},
	{
		Name:     "verify",
		Category: "offline mode",
		Usage:    "Verify the last randomness contributed to Phase 2",
		Action:   actions.Phase2Verify,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "phase2prev",
				Usage:    "Phase 2 file being an input to the contribution",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "phase2next",
				Usage:    "Phase 2 file that was contributed to",
				Required: true,
			},
		},
	},
	{
		Name:     "extract-keys",
		Category: "offline mode",
		Usage:    "Extract Proving and Verifying Keys",
		Action:   actions.Phase2ExtractKeys,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "r1cs",
				Usage:    "R1CS file generated from a gnark circuit",
				Required: true,
			},
			&cli.StringFlag{
				Name: "srscommons",
				Usage: "Circuit-independent components of the Groth16 SRS file generated on the Phase 2" +
					" initialization",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "beacon",
				Usage:    "Random string generated on the Phase 2 initialization",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name: "phase2",
				Usage: "List of Phase 2 files to verify the contributions in the order they were\n" +
					"created. Contributions are verified in pairs, so at least two files must be provided.\n" +
					" This DOES NOT INCLUDE the original Phase 2 file generated on initialization.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "pk",
				Usage:    "Output path for the proving key",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "vk",
				Usage:    "Output path for the verifying key",
				Required: true,
			},
		},
	},
}
