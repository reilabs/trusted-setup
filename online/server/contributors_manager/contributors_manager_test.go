package contributors_manager_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
)

func TestEnsure(t *testing.T) {
	coord := contributors_manager.New()

	_ = coord.AddContributor("a")
	assert.NoError(t, coord.EnsureNextContributor("a"))
	assert.Error(t, coord.EnsureNextContributor("b"))

	_ = coord.AddContributor("b")
	assert.NoError(t, coord.EnsureNextContributor("a"))
	assert.Error(t, coord.EnsureNextContributor("b"))
}

func TestRemove(t *testing.T) {
	coord := contributors_manager.New()

	_ = coord.AddContributor("a")
	assert.NoError(t, coord.EnsureNextContributor("a"))
	assert.Error(t, coord.EnsureNextContributor("b"))

	assert.NoError(t, coord.RemoveCurrentContributor())
	assert.Error(t, coord.EnsureNextContributor("a"))
}

func TestNotify(t *testing.T) {
	coord := contributors_manager.New()

	notifyA := coord.AddContributor("a")
	assert.NotNil(t, notifyA)
	notifyB := coord.AddContributor("b")
	assert.NotNil(t, notifyB)

	onAPositionUpdate := func(newPosition int) {
		assert.Equal(t, 0, newPosition)
		assert.NoError(t, coord.RemoveCurrentContributor())
	}
	go notifyA(onAPositionUpdate)

	onBPositionUpdate := func(newPosition int) {
		assert.Equal(t, 0, newPosition)
		assert.NoError(t, coord.RemoveCurrentContributor())
	}
	go notifyB(onBPositionUpdate)
}
