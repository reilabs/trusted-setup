// Package ceremony_service implements the gRPC service for the multi-party computation setup ceremony.
package ceremony_service

import (
	"context"
	"log"

	"google.golang.org/grpc/peer"

	"github.com/reilabs/trusted-setup/online/api"
	"github.com/reilabs/trusted-setup/online/api/stream_utils"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
)

type ceremonyService struct {
	api.UnimplementedCeremonyServiceServer

	name        string
	coordinator coordinator.Coordinator
}

// New returns a new instance of CeremonyServiceServer.
//
// # The returned object can be passed to the gRPC server constructor
//
// Accepts a name of the ceremony and a ceremony coordinator instance.
func New(
	name string, coordinator coordinator.Coordinator,
) api.CeremonyServiceServer {
	return &ceremonyService{name: name, coordinator: coordinator}
}

func clientAddressFromContext(ctx context.Context) string {
	peerInfo, ok := peer.FromContext(ctx)
	clientIP := "unknown"
	if ok && peerInfo.Addr != nil {
		clientIP = peerInfo.Addr.String()
	}

	return clientIP
}

func onContributorPositionUpdate(newPosition int, clientIp string, stream api.CeremonyService_ContributeServer) {
	log.Printf("contributor %s got slot %d in the queue", clientIp, newPosition)
	if err := stream.Send(api.NewTurnNotification(newPosition)); err != nil {
		log.Printf("failed to send position update to %s: %v", clientIp, err)
	}
}

// Contribute implements the flow of a single contribution coming from a contributor client.
func (s *ceremonyService) Contribute(
	stream api.CeremonyService_ContributeServer,
) error {
	err := stream.Send(api.NewHello(s.name))
	if err != nil {
		return err
	}

	clientIp := clientAddressFromContext(stream.Context())
	waitForThisContributorsTurn := s.coordinator.AddContributor(
		func(newPosition int) {
			onContributorPositionUpdate(newPosition, clientIp, stream)
		},
	)

	waitForThisContributorsTurn()

	log.Printf("Sending last contribution to %s", clientIp)
	n, err := s.coordinator.WriteLastContribution(stream_utils.NewStreamWriter(stream))
	if err != nil {
		log.Printf("error sending last contribution to %s", clientIp)
		return err
	}
	log.Printf("Sent %d bytes", n)

	log.Printf("Contribution to be received from %s", clientIp)
	n, err = s.coordinator.ReadNextContribution(stream_utils.NewStreamReader(stream))
	log.Printf("Received %d bytes", n)
	if err != nil {
		log.Printf("%s: %v", clientIp, err)
		return stream.Send(api.NewValidationResponse(err))
	}
	log.Printf("Contribution from %s accepted", clientIp)

	return stream.Send(api.NewValidationResponse(nil))
}
