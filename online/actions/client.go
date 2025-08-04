package actions

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/online/client"
)

func Client(_ context.Context, cmd *cli.Command) error {
	host := cmd.String("host")
	port := cmd.String("port")

	c, err := client.New(host, port)
	if err != nil {
		return err
	}

	return c.Contribute()
}
