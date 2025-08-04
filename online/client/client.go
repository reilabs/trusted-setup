// Package client provides a client for the ceremony service.
//
// The client is used by contributors to connect to the coordinator and contribute to the ceremony.
package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/reilabs/trusted-setup/online/api"
	"github.com/reilabs/trusted-setup/online/api/stream_utils"
	"github.com/reilabs/trusted-setup/online/phase2"
)

// ConnectAndContribute connects to the coordinator and contributes to the ceremony.
//
// The host and port parameters are the IP address and port of the coordinator.
// The function blocks when the contributor is waiting for their turn.
//
// After the client finishes contributing, the connection is closed and the function returns nil.
// The function returns an error if the connection to the coordinator fails or if the server
// rejects the connection or contribution.
func ConnectAndContribute(host string, port string) error {
	hostPort := host + ":" + port
	log.Printf("Connecting to %s...", hostPort)

	conn, err := grpc.NewClient(
		hostPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("failed to close connection: %v", err)
		}
	}(conn)

	client := api.NewCeremonyServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	joinResp, err := client.Join(ctx, &api.JoinRequest{Version: api.ProtocolVersion})
	if err != nil {
		return err
	}
	log.Printf("Joined ceremony: %s", joinResp.CeremonyName)
	if joinResp.IsAccepted {
		log.Print("Connection successful")
	} else {
		return fmt.Errorf("connection rejected: %s", joinResp.RejectionReason)
	}

	waitStream, err := client.WaitForTurn(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	log.Print("Contribution slot requested")
	var contribResponse *api.TurnNotification
	for {
		contribResponse, err = waitStream.Recv()
		if err != nil {
			log.Printf("error receiving from stream: %v", err)
			return err
		}
		log.Printf("Contribution slot assigned, position in queue: %d", contribResponse.PositionInQueue)

		if !contribResponse.CanContribute {
			log.Printf("Waiting for our turn...")
			continue
		}
		break
	}

	log.Print("Our turn, downloading last contribution")
	dlStream, err := client.DownloadContribution(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	p2 := phase2.NewContributor()
	downloader := stream_utils.NewStreamDownloader(dlStream)
	n, err := p2.ReadFrom(downloader)
	if err != nil {
		log.Fatalf("failed to download last phase 2: %v", err)
	}
	log.Printf("Received %d bytes", n)

	p2.Contribute()

	log.Print("Uploading our contribution")
	ulStream, err := client.UploadContribution(ctx)
	if err != nil {
		return err
	}

	uploader := stream_utils.NewStreamUploader(ulStream)
	n, err = p2.WriteTo(uploader)
	if err != nil {
		log.Fatalf("failed to upload next phase 2: %v", err)
	}
	log.Printf("Sent %d bytes", n)

	ulResp, err := ulStream.CloseAndRecv()
	if err != nil {
		return err
	}
	if !ulResp.IsValid {
		log.Fatalf("contribution rejected: %s", ulResp.RejectionReason)
	}
	log.Print("Contribution accepted")

	return nil
}
