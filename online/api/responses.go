package api

import (
	"fmt"
)

// NewHello returns a Hello message with the given ceremony name.
func NewHello(ceremonyName string) *ContributeResponse {
	return &ContributeResponse{
		Response: &ContributeResponse_Hello{
			Hello: &Hello{
				CeremonyName: ceremonyName,
			},
		},
	}
}

// NewDataChunk returns a DataChunk message with the given data.
func NewDataChunk(data []byte) *DataChunk {
	return &DataChunk{
		Data: data,
	}
}

// NewLastContribution returns a ContributeResponse_LastContribution message with the given data.
func NewLastContribution(data []byte) *ContributeResponse {
	return &ContributeResponse{
		Response: &ContributeResponse_LastContribution{
			LastContribution: NewDataChunk(data),
		},
	}
}

// NewTurnNotification returns a TurnNotification message with the given position.
func NewTurnNotification(position int) *ContributeResponse {
	return &ContributeResponse{
		Response: &ContributeResponse_Turn{
			Turn: &TurnNotification{
				CanContribute:   position == 0,
				PositionInQueue: uint32(position),
			},
		},
	}
}

// NewValidationResponse returns a ValidationResponse message with the given error.
//
// If the error is nil, the response is considered valid.
func NewValidationResponse(err error) *ContributeResponse {
	return &ContributeResponse{
		Response: &ContributeResponse_Validation{
			Validation: &ValidationResponse{
				IsValid:         err == nil,
				RejectionReason: fmt.Sprintf("%v", err),
			},
		},
	}
}
