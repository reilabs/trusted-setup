// Package ceremony_service implements the gRPC service for the multi-party computation setup ceremony.
package ceremony_service

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/peer"

	"github.com/reilabs/trusted-setup/online/api"
	"github.com/reilabs/trusted-setup/online/api/stream_utils"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
)

type ceremonyService struct {
	api.UnimplementedCeremonyServiceServer

	name        string
	coordinator coordinator.Coordinator
	log         *zerolog.Logger
}

// New returns a new instance of CeremonyServiceServer.
//
// Accepts a name of the ceremony, ceremony coordinator instance and a logger.
// The ceremony name is sent to contributors after they connect.
// The coordinator keeps track of incoming contributions, accepts new contribution
// candidates and validates them.
// The logger will accept log entries with crucial steps of the ceremony, that can
// be useful during attestation or keys recovery.
//
// The returned object can be passed to the gRPC server constructor
func New(
	name string, coordinator coordinator.Coordinator, log *zerolog.Logger,
) api.CeremonyServiceServer {
	log.Info().Str("name", name).Msg("new ceremony started")

	return &ceremonyService{name: name, coordinator: coordinator, log: log}
}

func clientAddressFromContext(ctx context.Context) string {
	peerInfo, ok := peer.FromContext(ctx)
	clientIP := "unknown"
	if ok && peerInfo.Addr != nil {
		clientIP = peerInfo.Addr.String()
	}

	return clientIP
}

// Contribute implements the flow of a single contribution coming from a contributor client.
func (s *ceremonyService) Contribute(
	stream api.CeremonyService_ContributeServer,
) error {
	clientIp := clientAddressFromContext(stream.Context())
	s.log.Info().
		Str("ip", clientIp).
		Msg("new contributor connected")

	err := stream.Send(api.NewHello(s.name))
	if err != nil {
		return err
	}

	waitForThisContributorsTurn := s.coordinator.AddContributor(
		func(newPosition int) {
			err = stream.Send(api.NewTurnNotification(newPosition))
			s.log.Info().
				Int("newQueuePosition", newPosition).
				Str("ip", clientIp).
				Err(err).
				Msg("contributor position update")
		},
	)

	waitForThisContributorsTurn()

	s.log.Info().
		Str("ip", clientIp).
		Msg("sending last accepted contribution")
	n, err := s.coordinator.WriteLastContribution(stream_utils.NewStreamWriter(stream))
	if err != nil {
		s.log.Info().
			Str("ip", clientIp).
			Err(err)
		return err
	}
	s.log.Info().
		Str("ip", clientIp).
		Int64("size", n).
		Msg("sent last accepted contribution")

	s.log.Info().
		Str("ip", clientIp).
		Msg("receiving new contribution candidate")
	n, err = s.coordinator.ReadNextContribution(stream_utils.NewStreamReader(stream))
	if err != nil {
		s.log.Info().
			Str("ip", clientIp).
			Int64("size", n).
			Err(err).
			Msg("new contribution candidate rejected")
		return stream.Send(api.NewValidationResponse(err))
	}

	s.log.Info().
		Str("ip", clientIp).
		Int64("size", n).
		Msg("new contribution candidate accepted")

	return stream.Send(api.NewValidationResponse(nil))
}
