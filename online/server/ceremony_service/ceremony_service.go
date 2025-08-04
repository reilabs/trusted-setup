package ceremony_service

import (
	"context"
	"log"

	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/reilabs/trusted-setup/online/api"
	"github.com/reilabs/trusted-setup/online/api/stream_utils"
	"github.com/reilabs/trusted-setup/online/phase2"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
)

type ceremonyService struct {
	api.UnimplementedCeremonyServiceServer

	name        string
	coordinator coordinator.Coordinator
}

// New returns a new instance of CeremonyServiceServer.
//
// name is a human-readable name of the ceremony sent to connected contributors. p2 implements Phase 2 management
// logic from the coordinator perspective. manager implements contributor management logic.
func New(
	name string, p2 phase2.Coordinator, manager contributors_manager.ContributorsManager,
) api.CeremonyServiceServer {
	return &ceremonyService{name: name, coordinator: coordinator.New(p2, manager)}
}

func clientAddressFromContext(ctx context.Context) string {
	peerInfo, ok := peer.FromContext(ctx)
	clientIP := "unknown"
	if ok && peerInfo.Addr != nil {
		clientIP = peerInfo.Addr.String()
	}

	return clientIP
}

// Join handles the Join protocol method defined in api/ceremony.proto.
//
// This is a method contributors call to join the ceremony. The supported protocol version is checked to make sure
// older clients are not allowed to talk to newer servers.
func (s *ceremonyService) Join(ctx context.Context, req *api.JoinRequest) (*api.JoinResponse, error) {
	clientIP := clientAddressFromContext(ctx)

	if req.Version != api.ProtocolVersion {
		log.Printf("contributor %s rejected: unsupported protocol (0x%04x)", clientIP, req.Version)
		return connectResponseVersionUnsupported(s.name, req.Version), nil
	}

	log.Printf("contributor connected: %s, protocol version: 0x%04x", clientIP, req.Version)
	return connectResponseOk(s.name), nil
}

// WaitForTurn handles the WaitForTurn protocol method defined in api/ceremony.proto.
//
// Contributors call this method to get a slot to contribute to the ceremony.
//
// The service creates a new contributor out of the incoming stream. The first contributor
// from the queue is then notified to start the contribution while all the others wait.
//
// The method blocks for every contributor. It unblocks only after the contributor submits
// their contribution or if the client disconnects.
//
// TODO when client disconnects without contributing (but after getting the slot), other
// waiting contributors will be blocked forever.
func (s *ceremonyService) WaitForTurn(
	_ *emptypb.Empty, stream api.CeremonyService_WaitForTurnServer,
) error {
	clientIp := clientAddressFromContext(stream.Context())

	notifyPositionUpdate := s.coordinator.AddContributor(clientIp)

	notifyPositionUpdate(
		func(newPosition int) {
			log.Printf("contributor %s got slot %d in the queue", clientIp, newPosition)
			if err := stream.Send(turnNotification(newPosition)); err != nil {
				log.Printf("failed to send position update to %s: %v", clientIp, err)
			}
		},
	)

	return nil
}

// DownloadContribution handles the DownloadContribution protocol method defined in api/ceremony.proto.
//
// Contributors call this method to get the last verified contribution from the server.
func (s *ceremonyService) DownloadContribution(
	_ *emptypb.Empty, stream api.CeremonyService_DownloadContributionServer,
) error {
	clientIp := clientAddressFromContext(stream.Context())
	log.Printf("Last contribution requested from %s", clientIp)

	n, err := s.coordinator.WriteLastContribution(clientIp, stream_utils.NewStreamUploader(stream))
	if err != nil {
		return err
	}
	log.Printf("Sent %d bytes", n)

	return nil
}

// UploadContribution handles the UploadContribution protocol method defined in api/ceremony.proto.
//
// Contributors call this method to sent their contribution to the server for verification.
func (s *ceremonyService) UploadContribution(stream api.CeremonyService_UploadContributionServer) error {
	clientIp := clientAddressFromContext(stream.Context())
	log.Printf("Contribution to be received from %s", clientIp)

	n, err := s.coordinator.ReadNextContribution(clientIp, stream_utils.NewStreamDownloader(stream))
	log.Printf("Received %d bytes", n)
	if err != nil {
		return stream.SendAndClose(uploadResponseFailed(err))
	}

	err = s.coordinator.VerifyNextContribution()
	if err != nil {
		log.Printf("Contribution from %s rejected: %v", clientIp, err)
		return err
	}

	log.Printf("Contribution from %s accepted", clientIp)
	return stream.SendAndClose(uploadResponseOk())
}
