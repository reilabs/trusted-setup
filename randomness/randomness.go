// Package randomness provides "publicly verifiable, unbiased and unpredictable random values" as guaranteed
// by the drand organization (https://github.com/drand/drand).
package randomness

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/drand/go-clients/client"
	"github.com/drand/go-clients/client/http"
)

// Default network chain hash as per the drand project documentation.
const chainHash = "8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce"

// API returning the randomness, as per the drand project documentation.
const apiHost = "https://api.drand.sh/"

// Beacon contains 32 bytes of randomness acquired at the module initialization. The value does not change within
// the single program session.
var beacon []byte = nil

// init initializes the randomness module. It connects to the drand network and gets a random value.
func init() {
	httpClient, err := http.NewSimpleClient(apiHost, chainHash)
	if err != nil {
		panic(err)
	}
	chb, err := hex.DecodeString(chainHash)
	if err != nil {
		panic(err)
	}

	c, err := client.New(
		client.From(httpClient),
		client.WithChainHash(chb),
	)
	if err != nil {
		panic(err)
	}

	const mostRecentKnownRound = 0
	r, err := c.Get(context.Background(), mostRecentKnownRound)
	if err != nil {
		panic(err)
	}

	beacon = r.GetRandomness()
	if len(beacon) != 32 {
		panic(fmt.Errorf("randomness: expected 32 bytes, got %d", len(beacon)))
	}
	if beacon == nil {
		panic(fmt.Errorf("randomness: drand did not return randomness"))
	}
}

// GetBeacon returns the 32 bytes of randomness acquired at the module initialization.
func GetBeacon() []byte {
	return beacon
}
