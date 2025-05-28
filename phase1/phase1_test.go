package phase1

import (
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

type dummyCircuit struct {
	X frontend.Variable
	Y frontend.Variable
}

func (c *dummyCircuit) Define(api frontend.API) error {
	api.AssertIsEqual(c.Y, api.Mul(c.X, c.X))
	return nil
}

func setupR1CS() (string, *os.File, error) {
	circuit := &dummyCircuit{}
	cs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		return "", nil, err
	}

	r1csFile, err := os.CreateTemp("", "test.*.cs")
	if err != nil {
		return "", nil, err
	}

	_, err = cs.WriteTo(r1csFile)
	if err != nil {
		return "", nil, err
	}

	return r1csFile.Name(), r1csFile, nil
}

func teardown(files ...*os.File) {
	for _, file := range files {
		if file != nil {
			_ = os.Remove(file.Name())
			_ = file.Close()
		}
	}
}

func TestPhase1FromPtau(t *testing.T) {
	r1csFilePath, r1csFile, err := setupR1CS()
	if err != nil {
		t.Fatalf("failed during setup: %v", err)
	}
	defer teardown(r1csFile)

	phase2File, err := os.CreateTemp("", "test.*.phase2")
	if err != nil {
		t.Fatalf("failed to create phase2 file: %v", err)
	}
	defer teardown(phase2File)

	ptauFilePath := "phase1_test.ptau"

	if err := FromPtauPath(ptauFilePath, r1csFilePath, phase2File.Name()); err != nil {
		t.Errorf("phase1FromPtau returned an error: %v", err)
	}
}
