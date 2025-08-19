package contribution_test

import (
	"bytes"
	"testing"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/offline/phase1"
	"github.com/reilabs/trusted-setup/offline/r1cs"
	"github.com/reilabs/trusted-setup/online/contribution"
	test_circuit "github.com/reilabs/trusted-setup/test"
)

func setup() (*cs.R1CS, *mpcsetup.Phase1, []byte) {
	ccs, err := r1cs.FromFile("../test/resources/server.r1cs")
	if err != nil {
		panic(err)
	}

	p1, err := phase1.FromFile("../test/resources/server.ph1")
	if err != nil {
		panic(err)
	}

	return ccs, p1, bytes.Repeat([]byte{0x42}, 32)
}

func teardown(ccs *cs.R1CS, pk *groth16.ProvingKey, vk *groth16.VerifyingKey) {
	err := test_circuit.ProveAndVerify(ccs, pk, vk)
	if err != nil {
		panic(err)
	}
}

func Test(t *testing.T) {
	// Generate initial data: constraint system, Phase 1 and random beacon
	ccs, p1, beacon := setup()

	// Initialize Phase 2 from Phase 1, circuit constraint system and random beacon
	p2 := contribution.New(p1, ccs, beacon)
	assert.Equal(t, 0, p2.GetCount())

	// Serialize initial Phase 2 to a buffer
	var buf bytes.Buffer
	_, err := p2.WriteTo(&buf)
	assert.NoError(t, err)

	// Recreate the initial contribution from a buffer
	contrib := contribution.NewContributable()
	_, err = contrib.ReadFrom(&buf)
	assert.NoError(t, err)

	// Contribute
	contrib.Contribute()

	// Submit contribution
	err = p2.AddContribution(contrib.(contribution.Verifiable))
	assert.NoError(t, err)
	assert.Equal(t, 1, p2.GetCount())

	// One contribution should be enough to generate keys
	pk, vk := p2.ExtractKeys()

	// Check that keys can be used for proof generation and verification
	teardown(ccs, &pk, &vk)
}
