package coordinator

import (
	"io"
	"log"

	"github.com/reilabs/trusted-setup/online/phase2"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
)

// Coordinator is an interface for managing contributors and their contributions during the ceremony.
type Coordinator interface {
	AddContributor(clientId string) contributors_manager.PositionUpdateNotifier
	WriteLastContribution(clientId string, client io.Writer) (int64, error)
	ReadNextContribution(clientId string, client io.Reader) (int64, error)
	VerifyNextContribution() error
}

type coordinator struct {
	last    phase2.Coordinator
	next    phase2.Verifiable
	manager contributors_manager.ContributorsManager
}

// New creates a new coordinator object.
//
// p2 implements Phase 2 management logic from the coordinator perspective.
// manager implements contributor management logic.
func New(
	p2 phase2.Coordinator, manager contributors_manager.ContributorsManager,
) Coordinator {
	return &coordinator{
		last:    p2,
		next:    p2.NewVerifiable(),
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
func (s *coordinator) AddContributor(clientId string) contributors_manager.PositionUpdateNotifier {
	return s.manager.AddContributor(clientId)
}

// WriteLastContribution writes the last known contribution to the given contributor's writer.
//
// Before writing, it checks if the given clientId identifies the next eligible contributor.
// If the ID does not match, the function returns an error.
//
// Returns the number of bytes written to the contributor and error (if happened).
func (s *coordinator) WriteLastContribution(
	clientId string, contributor io.Writer,
) (int64, error) {
	if err := s.manager.EnsureNextContributor(clientId); err != nil {
		log.Printf("%s is not eligible for contributing", clientId)
		return 0, err
	}

	return s.last.WriteTo(contributor)
}

// ReadNextContribution reads the next upcoming contribution from the given contributor's reader.
//
// Before reading, it checks if the given clientId identifies the next eligible contributor.
// If the ID does not match, the function returns an error.
//
// Returns the number of bytes read from the contributor and error (if happened).
func (s *coordinator) ReadNextContribution(clientId string, contributor io.Reader) (int64, error) {
	if err := s.manager.EnsureNextContributor(clientId); err != nil {
		log.Printf("%s is not eligible for contributing", clientId)
		return 0, err
	}

	return s.next.ReadFrom(contributor)
}

// VerifyNextContribution verifies if the contribution loaded with ReadNextContribution is valid.
//
// If the verification is positive, the new contribution is considered the last verified
// and will be handed out to the next contributor in the queue.
//
// If the verification fails, the last verified contribution remains unchanged.
//
// Regardless the result, the contributor removes the current contributor for the queue,
// allowing the next one to contribute.
func (s *coordinator) VerifyNextContribution() error {
	if err := s.last.AddContribution(s.next); err != nil {
		if removeErr := s.manager.RemoveCurrentContributor(); removeErr != nil {
			log.Printf("cannot remove current contributor: %v", removeErr)
		}
		return err
	}

	if removeErr := s.manager.RemoveCurrentContributor(); removeErr != nil {
		log.Printf("cannot remove current contributor: %v", removeErr)
	}

	return nil
}
