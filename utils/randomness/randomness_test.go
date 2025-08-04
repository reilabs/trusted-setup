package randomness_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/utils/randomness"
)

func Test(t *testing.T) {
	r, err := randomness.New()
	assert.NoError(t, err)

	got := r.GetBeacon()
	assert.NotEmpty(t, got)
	assert.Equal(t, 32, len(got))
}
