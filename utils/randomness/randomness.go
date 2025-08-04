// Package randomness provides "publicly verifiable, unbiased and unpredictable random values" as guaranteed
// by the drand organization (https://github.com/drand/drand).
package randomness

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/drand/go-clients/client"
	"github.com/drand/go-clients/client/http"
	"github.com/drand/go-clients/drand"
)

// BeaconProvider defines an interface to get 32 bytes of randomness (beacon).
type BeaconProvider interface {
	GetBeacon() []byte
}

// drandProvider implements BeaconProvider by retrieving randomness from drand network.
type drandProvider struct {
	client drand.Client
}

// New initializes the randomness module. It connects to the drand network, so random value can be
// obtained with GetBeacon.
func New() (BeaconProvider, error) {
	// Default network chain hash as per the drand project documentation.
	const chainHash = "8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce"
	// API returning the randomness, as per the drand project documentation.
	const apiHost = "https://api.drand.sh/"

	httpClient, err := http.NewSimpleClient(apiHost, chainHash)
	if err != nil {
		panic(err)
	}
	chb, err := hex.DecodeString(chainHash)
	if err != nil {
		panic(err)
	}

	p := drandProvider{}
	p.client, err = client.New(
		client.From(httpClient),
		client.WithChainHash(chb),
	)
	if err != nil {
		panic(err)
	}

	return &p, nil
}

// GetBeacon returns the 32 bytes of randomness.
func (d *drandProvider) GetBeacon() []byte {
	const mostRecentKnownRound = 0
	r, err := d.client.Get(context.Background(), mostRecentKnownRound)
	if err != nil {
		panic(err)
	}

	beacon := r.GetRandomness()
	if len(beacon) != 32 {
		panic(fmt.Errorf("randomness: expected 32 bytes, got %d", len(beacon)))
	}
	if beacon == nil {
		panic(fmt.Errorf("randomness: drand did not return randomness"))
	}

	return beacon
}
