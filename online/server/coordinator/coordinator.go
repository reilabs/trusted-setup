// Package coordinator implements the coordinator logic for the multi-party computation setup ceremony.
// Coordinator manages contributors and their contributions during the ceremony.
package coordinator

import (
	"fmt"
	"io"
	"log"

	"github.com/reilabs/trusted-setup/online/contribution"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
)

// Coordinator is an interface for managing contributors and their contributions during the ceremony.
type Coordinator interface {
	AddContributor(notify contributors_manager.OnPositionUpdate) contributors_manager.PositionUpdateNotifier
	WriteLastContribution(client io.Writer) (int64, error)
	ReadNextContribution(client io.Reader) (int64, error)
}

type coordinator struct {
	last    contribution.Contribution
	manager contributors_manager.ContributorsManager
}

// New creates a new coordinator object.
//
// p2 implements Phase 2 management logic from the coordinator perspective.
// manager implements contributor management logic.
func New(
	p2 contribution.Contribution, manager contributors_manager.ContributorsManager,
) Coordinator {
	return &coordinator{
		last:    p2,
		manager: manager,
	}
}

// AddContributor adds a new contributor to the ceremony.
//
// clientId is a string identifying contributor in a ceremony.
//
// Returns a PositionUpdateNotifier function to be called by the caller. The function blocks, waiting
// for updates of the added contributor's position in the queue. Whenever there is an update, the callback
// being and argument to PositionUpdateNotifier is called.
func (s *coordinator) AddContributor(notify contributors_manager.OnPositionUpdate) contributors_manager.PositionUpdateNotifier {
	return s.manager.AddContributor(notify)
}

// WriteLastContribution writes the last known contribution to the given contributor's writer.
//
// Returns the number of bytes written to the contributor and error (if happened).
// In case of error, the current contributor is removed from the ceremony.
func (s *coordinator) WriteLastContribution(
	contributor io.Writer,
) (int64, error) {
	n, err := s.last.WriteTo(contributor)
	if err != nil {
		if err := s.manager.RemoveCurrentContributor(); err != nil {
			log.Printf("error removing current contributor: %v", err)
		}
	}
	return n, err
}

// ReadNextContribution reads and verifies the next upcoming contribution from the given contributor's reader.
//
// If the verification is positive, the new contribution is considered the last verified
// and will be handed out to the next contributor in the queue.
//
// Returns the number of bytes read from the contributor and error (if happened).
// Upon return, regardless if any error happened, the current contributor is removed from the ceremony.
func (s *coordinator) ReadNextContribution(contributor io.Reader) (int64, error) {
	next := s.last.NewVerifiable()
	n, err := next.ReadFrom(contributor)
	if err != nil {
		if err := s.manager.RemoveCurrentContributor(); err != nil {
			log.Printf("error removing current contributor: %v", err)
		}
		return n, fmt.Errorf("error reading next contribution: %w", err)
	}

	err = s.last.AddContribution(next)

	if err := s.manager.RemoveCurrentContributor(); err != nil {
		log.Printf("error removing current contributor: %v", err)
	}

	if err != nil {
		return n, fmt.Errorf("error verifying next contribution: %w", err)
	}

	return n, nil
}
