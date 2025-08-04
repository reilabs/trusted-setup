package contributors_manager_test

import (
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
)

func TestNotify(t *testing.T) {
	cm := contributors_manager.New()

	aHandlerCalledCount := 0
	onAPositionUpdate := func(newPosition int) {
		aHandlerCalledCount++
		log.Printf("Contributor A position updated to %d", newPosition)
		assert.Equal(t, 0, newPosition)
		assert.NoError(t, cm.RemoveCurrentContributor())
	}

	bExpectedPosition := 1
	bHandlerCalledCount := 0
	var wg sync.WaitGroup
	wg.Add(2)
	onBPositionUpdate := func(newPosition int) {
		defer wg.Done()
		bHandlerCalledCount++
		log.Printf("Contributor B position updated to %d", newPosition)
		assert.Equal(t, bExpectedPosition, newPosition)
		bExpectedPosition -= 1
	}

	notifyA := cm.AddContributor(onAPositionUpdate)
	assert.NotNil(t, notifyA)
	notifyB := cm.AddContributor(onBPositionUpdate)
	assert.NotNil(t, notifyB)

	go notifyB() // run in background because it will block waiting for A
	notifyA()

	wg.Wait()
	assert.Equal(t, 1, aHandlerCalledCount) // Only learned about position 0
	assert.Equal(t, 2, bHandlerCalledCount) // Learned about position 1, then update to 0

}
