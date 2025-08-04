// Package contributors_manager implements a manager for contributors in the multi-party computation setup ceremony.
// Contributors Manager handles the queue of the contributors. It allows for adding new contributors to the ceremony,
// removing the current contributor after they've been handled and notifying other contributors about their position
// in the queue.
package contributors_manager

import (
	"github.com/reilabs/trusted-setup/online/server/contributors_manager/unique_fifo"
)

type contributor struct {
	positionChannel chan int
}

func newContributor() contributor {
	return contributor{make(chan int, 1)}
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
// It blocks, calling OnPositionUpdate repeatedly on every update of the contributor's position in the queue.
// It returns when there will be no more updates.
type PositionUpdateNotifier func()

type ContributorsManager interface {
	AddContributor(notify OnPositionUpdate) PositionUpdateNotifier
	RemoveCurrentContributor() error
}

// New returns a new instance of ContributorsManager.
func New() ContributorsManager {
	return &manager{
		contributorsQueue: unique_fifo.New[contributor](),
	}
}

// AddContributor adds a new contributor to the queue.
//
// It returns a PositionUpdateNotifier that can be used to get notified about the updates.
// PositionUpdateNotifier blocks, calling OnPositionUpdate repeatedly on every update of the contributor's position in the queue.
// It returns when there will be no more updates.
//
// Example:
//
//	cm := contributors_manager.New()
//	onUpdate := notifier(func(position int) {
//		log.Printf("contributor %s is at position %d", ip, position)
//	})
//	notifier := cm.AddContributor(onUpdate)
//	notifier() // blocks, calling onUpdate() on contributor's position update, until contributor reaches position 0.

func (c *manager) AddContributor(notify OnPositionUpdate) PositionUpdateNotifier {
	nc := newContributor()
	ncPosition := c.contributorsQueue.Enqueue(nc)
	nc.positionChannel <- ncPosition
	return func() {
		for positionUpdate := range nc.positionChannel {
			notify(positionUpdate)
			if positionUpdate == 0 {
				return
			}
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

	return c.notifyWaitingContributors()
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
		wc.positionChannel <- i
	}

	return nil
}
