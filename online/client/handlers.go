package client

import (
	"fmt"
	"io"
	"log"

	"github.com/reilabs/trusted-setup/online/api"
	"github.com/reilabs/trusted-setup/online/api/stream_utils"
	"github.com/reilabs/trusted-setup/online/contribution"
)

func (c *Client) messageLoop() error {
	for {
		resp, err := c.stream.Recv()
		if err == io.EOF {
			log.Print("Contribution stream closed")
			break
		}
		if err != nil {
			log.Printf("error receiving from stream: %v", err)
			return err
		}
		switch r := resp.Response.(type) {
		case *api.ContributeResponse_Hello:
			c.onHello(r)
		case *api.ContributeResponse_Turn:
			c.onTurn(r)
		case *api.ContributeResponse_Validation:
			return c.onValidation(r)
		default:
			log.Printf("unexpected response type: %T", r)
		}
	}

	return nil
}

func (c *Client) onHello(msg *api.ContributeResponse_Hello) {
	log.Printf("Joined ceremony: %s", msg.Hello.CeremonyName)
}

func (c *Client) onTurn(msg *api.ContributeResponse_Turn) {
	log.Printf("Contribution slot assigned, position in queue: %d", msg.Turn.PositionInQueue)
	if !msg.Turn.CanContribute {
		log.Printf("Waiting for our turn...")
		return
	}

	c.contribute()
}

func (c *Client) onValidation(msg *api.ContributeResponse_Validation) error {
	if !msg.Validation.IsValid {
		return fmt.Errorf("contribution rejected: %s", msg.Validation.RejectionReason)
	}

	log.Print("Contribution accepted")
	return nil
}

func (c *Client) contribute() {
	log.Print("Our turn, downloading last contribution")

	p2 := contribution.NewContributable()
	downloader := stream_utils.NewStreamReader(c.stream)
	n, err := p2.ReadFrom(downloader)
	if err != nil {
		log.Fatalf("failed to download last phase 2: %v", err)
	}
	log.Printf("Received %d bytes", n)

	log.Print("Generating contribution")
	p2.Contribute()

	log.Print("Uploading our contribution")
	uploader := stream_utils.NewStreamWriter(c.stream)
	n, err = p2.WriteTo(uploader)
	if err != nil {
		log.Fatalf("failed to upload next phase 2: %v", err)
	}
	log.Printf("Sent %d bytes", n)
}
