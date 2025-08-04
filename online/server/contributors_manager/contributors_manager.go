package contributors_manager

import (
	"errors"
	"log"

	"github.com/reilabs/trusted-setup/online/server/contributors_manager/unique_fifo"
)

type contributor struct {
	ip              string
	positionChannel chan int
}

type manager struct {
	contributorsQueue unique_fifo.Queue[contributor]
}

// OnPositionUpdate is a type of function that is called whenever a contributor's position in the queue
// is updated.
//
// It accepts the current position of the contributor in the queue. It is used by the user
// of ContributorsManager to process the updates.
type OnPositionUpdate func(int)

// PositionUpdateNotifier is a type of function returned by AddContributor when a new contributor is added.
//
// It accepts an OnPositionUpdate function. It is used by the user of ContributorsManager to get notified about
// the updates.
type PositionUpdateNotifier func(OnPositionUpdate)

type ContributorsManager interface {
	AddContributor(contributorIp string) PositionUpdateNotifier
	RemoveCurrentContributor() error
	EnsureNextContributor(candidateIp string) error
}

// New returns a new instance of ContributorsManager.
func New() ContributorsManager {
	return &manager{
		contributorsQueue: unique_fifo.NewUniqueFifo[contributor](),
	}
}

var (
	ErrContributorOutOfOrder = errors.New("contributor is not eligible for contribution")
)

// AddContributor adds a new contributor to the queue.
//
// It accepts the IP address of the contributor to be treated as a contributor's identifier.
//
// It returns a PositionUpdateNotifier that can be used to get notified about the updates.
//
// Example:
//
//	contributorManager := contributors_manager.New()
//	ip := "192.168.1.1:65498"
//	notifier := contributorManager.AddContributor(ip)
//	notifier(func(position int) {
//		log.Printf("contributor %s is at position %d", ip, position)
//	})
func (c *manager) AddContributor(contributorIp string) PositionUpdateNotifier {
	nc := contributor{contributorIp, make(chan int, 1)}
	ncPosition := c.contributorsQueue.Enqueue(nc)
	nc.positionChannel <- ncPosition
	return func(notify OnPositionUpdate) {
		for positionUpdate := range nc.positionChannel {
			notify(positionUpdate)
		}
	}
}

// RemoveCurrentContributor removes the current contributor from the queue.
//
// It returns an error if the queue is empty.
func (c *manager) RemoveCurrentContributor() error {
	cc, err := c.contributorsQueue.Dequeue()
	if err != nil {
		return err
	}
	close(cc.positionChannel)
	if err = c.notifyWaitingContributors(); err != nil {
		log.Printf("cannot notify contributors: %v", err)
	}

	return nil
}

func (c *manager) notifyWaitingContributors() error {
	if c.contributorsQueue.Len() == 0 {
		return nil
	}

	waitingContributors, err := c.contributorsQueue.PeekAll()
	if err != nil {
		return err
	}

	for i, wc := range waitingContributors {
		log.Printf("notifying client %s of position %d", wc.ip, i)
		wc.positionChannel <- i
	}

	return nil
}

// EnsureNextContributor check is the contributor described by the given IP is eligible for contribution in the ceremony
// at the given moment.
//
// It returns an ErrContributorOutOfOrder if the contributor is not eligible for contribution and nil otherwise.
// It returns an error if there are no contributors in the queue.
func (c *manager) EnsureNextContributor(candidateIp string) error {
	nextContributor, err := c.contributorsQueue.Peek()
	if err != nil {
		return err
	}

	log.Printf("Next eligible contributor: %s, comparing with %s", nextContributor.ip, candidateIp)
	if nextContributor.ip != candidateIp {
		return ErrContributorOutOfOrder
	}

	return nil
}
