package randomness_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/utils/randomness"
)

func TestGetBeaconWithMock(t *testing.T) {
	mockValue := bytes.Repeat([]byte{0x42}, 32)
	mock := &randomness.MockProvider{Beacon: mockValue}

	got := mock.GetBeacon()
	if !bytes.Equal(got, mockValue) {
		t.Fatalf("GetBeacon() = %x; want %x", got, mockValue)
	}
}

func TestGetBeaconWithDrand(t *testing.T) {
	drand, err := randomness.NewDrandProvider()
	assert.NoError(t, err)

	got := drand.GetBeacon()
	assert.NotEmpty(t, got)
	assert.Equal(t, 32, len(got))
}
