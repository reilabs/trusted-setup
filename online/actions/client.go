package actions

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/reilabs/trusted-setup/online/client"
)

func Client(_ context.Context, cmd *cli.Command) error {
	host := cmd.String("host")
	port := cmd.String("port")

	return client.ConnectAndContribute(host, port)
}
