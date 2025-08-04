package ceremony_service

import (
	"fmt"

	"github.com/reilabs/trusted-setup/online/api"
)

func connectResponseOk(name string) *api.JoinResponse {
	return &api.JoinResponse{
		CeremonyName: name,
		IsAccepted:   true,
	}
}

func connectResponseVersionUnsupported(name string, clientVersion uint32) *api.JoinResponse {
	return &api.JoinResponse{
		IsAccepted:   false,
		CeremonyName: name,
		RejectionReason: fmt.Sprintf(
			"Protocol version 0x%04x not supported. Supported version: 0x%04x", clientVersion, api.ProtocolVersion,
		),
	}
}

func uploadResponseFailed(reason error) *api.UploadResponse {
	return &api.UploadResponse{
		IsValid:         false,
		RejectionReason: fmt.Sprintf("%s", reason),
	}
}

func uploadResponseOk() *api.UploadResponse {
	return &api.UploadResponse{
		IsValid: true,
	}
}

func turnNotification(position int) *api.TurnNotification {
	return &api.TurnNotification{
		CanContribute:   position == 0,
		PositionInQueue: uint32(position),
	}
}
