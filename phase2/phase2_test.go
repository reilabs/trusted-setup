package phase2

import (
	"log"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

const r1csFilePath = "test.r1cs"
const phase2FilePath = "test.ph2"
const ptauFilePath = "test.ptau"

func setupR1CS() error {
	circuit := &dummyCircuit{}
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		return err
	}

	r1csFile, err := os.Create(r1csFilePath)
	if err != nil {
		return err
	}

	_, err = cs.WriteTo(r1csFile)
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	err := setupR1CS()
	if err != nil {
		log.Fatalf("failed during setup: %v", err)
	}

	// Run the tests
	exitVal := m.Run()

	os.Remove(r1csFilePath)
	os.Remove(phase2FilePath)
	os.Exit(exitVal)
}

type dummyCircuit struct {
	X frontend.Variable
	Y frontend.Variable
}

func (c *dummyCircuit) Define(api frontend.API) error {
	api.AssertIsEqual(c.Y, api.Mul(c.X, c.X))
	return nil
}

func TestPhase2FromPtauPath(t *testing.T) {
	if err := FromPtauPath(ptauFilePath, r1csFilePath, phase2FilePath); err != nil {
		t.Errorf("phase1FromPtau returned an error: %v", err)
	}
}
