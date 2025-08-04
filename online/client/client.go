// Package client provides a client for the ceremony service.
package client

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/reilabs/trusted-setup/online/api"
)

// Client is used by contributors to connect to the coordinator and contribute to the ceremony.
type Client struct {
	connection *grpc.ClientConn
	stream     api.CeremonyService_ContributeClient
}

// New creates a new Client for the trusted ceremony.
//
// The client immediately connects to the given host and port.
//
// On success, a Client object is returned. On failure, an error is returned.
func New(host string, port string) (*Client, error) {
	hostPort := host + ":" + port
	log.Printf("Connecting to %s...", hostPort)

	c := Client{}
	var err error
	c.connection, err = grpc.NewClient(
		hostPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := api.NewCeremonyServiceClient(c.connection)
	c.stream, err = client.Contribute(context.Background())
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Contribute contributes to the ceremony.
//
// On success, nil is returned. On failure, an error is returned.
func (c *Client) Contribute() error {
	defer func() {
		err := c.connection.Close()
		if err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	return c.messageLoop()
}
