package online

import (
	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/online/actions"
)

// Commands defines all CLI commands related to online operations.
var Commands = []*cli.Command{
	{
		Name:     "server",
		Category: "online mode",
		Usage:    "Start a Ceremony server",
		Action:   actions.Server,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Usage:    "JSON file containing the server configuration",
				Required: true,
			},
		},
	},
	{
		Name:     "client",
		Category: "online mode",
		Usage:    "Connect to a Ceremony server and provide contributions",
		Action:   actions.Client,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "host",
				Usage:    "address of the Ceremony server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "port",
				Usage:    "port the Ceremony server listens on",
				Required: true,
			},
		},
	},
}
