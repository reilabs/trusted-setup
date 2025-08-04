package online_test

import (
	"bytes"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/offline/phase1"
	"github.com/reilabs/trusted-setup/offline/r1cs"
	"github.com/reilabs/trusted-setup/online/client"
	server_config "github.com/reilabs/trusted-setup/online/config"
	"github.com/reilabs/trusted-setup/online/contribution"
	"github.com/reilabs/trusted-setup/online/server"
	"github.com/reilabs/trusted-setup/online/server/ceremony_service"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
	test_circuit "github.com/reilabs/trusted-setup/test"
)

// testOnlineCeremony verifies the trusted setup ceremony in the client/server mode.
//
// The whole ceremony is done automatically. The test logic only orchestrates the ceremony
// server and clients startup. Then, the whole ceremony, from converting ptau to Phase 1,
// through Phase 2 initialization, contribution and verification is done without supervision
// by the server and clients. In the end, keys are extracted, proof is created and verified.
func TestOnlineCeremony(t *testing.T) {
	t.Run("Start server", testStartServer)
	t.Run("Run contributions", testRunContributions)
	t.Run("Stop server", testStopServer)
	t.Run("Extract keys, prove and verify", testProveAndVerifyOnline)
}

var serv *server.CeremonyServer
var config *server_config.Config
var last contribution.Contribution

func testStartServer(t *testing.T) {
	var err error

	config, err = server_config.New("resources/config.json")
	assert.NoError(t, err)

	ccs, err := r1cs.FromFile(config.R1cs)
	assert.NoError(t, err)

	p1, err := phase1.FromFile(config.Phase1)
	assert.NoError(t, err)

	beacon := bytes.Repeat([]byte{0x42}, 32)

	last = contribution.New(p1, ccs, beacon)

	service := ceremony_service.New(config.CeremonyName, coordinator.New(last, contributors_manager.New()))

	serv = server.New(service)

	assert.NoError(t, serv.Start(config.Host, config.Port))
}

func testRunContributions(t *testing.T) {
	const clientsCount = 50
	// Run some clients synchronously to simulate contributors connecting slowly one by one
	for i := 0; i < clientsCount; i++ {
		c, err := client.New(config.Host, strconv.Itoa(config.Port))
		assert.NoError(t, err)
		assert.NoError(t, c.Contribute())
	}

	// Now run a group of clients simultaneously to simulate a more real-life case of multiple
	// contributors connecting randomly and waiting for their turn
	var wg sync.WaitGroup
	errCh := make(chan error, clientsCount)
	for i := 0; i < clientsCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := client.New(config.Host, strconv.Itoa(config.Port))
			if err != nil {
				errCh <- err
				return
			}
			errCh <- c.Contribute()
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err)
	}
}

func testStopServer(t *testing.T) {
	serv.Stop()
}

func testProveAndVerifyOnline(t *testing.T) {
	ccs, err := r1cs.FromFile(config.R1cs)
	assert.NoError(t, err)

	pk, vk := last.ExtractKeys()
	err = test_circuit.ProveAndVerify(ccs, &pk, &vk)
	assert.NoError(t, err)
}
