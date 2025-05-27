package test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	native_mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	gnark_r1cs "github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/keys"
	"github.com/reilabs/trusted-setup/phase1"
	"github.com/reilabs/trusted-setup/phase2"
	"github.com/reilabs/trusted-setup/r1cs"
)

func TestIntegration(t *testing.T) {
	setup()

	t.Run("Ptau", testPtau)
	t.Run("Init", testInit)
	t.Run("Contribute", testContribute)
	t.Run("Verify", testVerify)
	t.Run("Extract keys", testExtractKeys)

	teardown()
}

const phase1FileName = "test.phase1"
const phase2FileName = "test.phase2"
const phase2evalFileName = "test.eval"
const r1csFileName = "test.r1cs"
const pkFileName = "test.pk"
const vkFileName = "test.vk"

var phase2Contributed []string

type TestCircuit struct {
	PreImage frontend.Variable
	Hash     frontend.Variable `gnark:",public"`
}

func (circuit *TestCircuit) Define(api frontend.API) error {
	mimc, _ := mimc.NewMiMC(api)
	mimc.Write(circuit.PreImage)
	api.AssertIsEqual(circuit.Hash, mimc.Sum())

	return nil
}

func setup() {
	circuit := &TestCircuit{}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), gnark_r1cs.NewBuilder, circuit)
	if err != nil {
		panic(err)
	}
	err = r1cs.ToFile(ccs, r1csFileName)
	if err != nil {
		panic(err)
	}
}

func teardown() {
	filesToRemove := []string{
		phase1FileName,
		phase2FileName,
		phase2FileName + ".*",
		phase2evalFileName,
		r1csFileName,
		pkFileName,
		vkFileName,
	}

	for _, fileName := range filesToRemove {
		files, err := filepath.Glob(fileName)
		if err == nil {
			for _, file := range files {
				log.Printf("Removing %s", file)
				err = os.Remove(file)
				if err != nil {
					log.Printf("Error removing %s: %v", file, err)
				}
			}
		}
	}
}

func testPtau(t *testing.T) {
	const ptauFileName = "test.ptau"

	assert.NoError(t, phase1.FromPtau(ptauFileName, phase1FileName))

	p1, err := phase1.FromFile(phase1FileName)
	assert.NoError(t, err)

	// Check some fields of phase1 to make sure something is there
	assert.NotEmpty(t, p1.Parameters.G1.Tau)
	assert.NotEmpty(t, p1.Parameters.G1.AlphaTau)
	assert.NotEmpty(t, p1.Parameters.G1.BetaTau)
	assert.NotEmpty(t, p1.Parameters.G2.Tau)
	assert.NotEmpty(t, p1.Parameters.G2.Beta)

	assert.NotEmpty(t, p1.Hash)
}

func testInit(t *testing.T) {
	assert.NoError(t, phase2.Init(phase1FileName, r1csFileName, phase2FileName, phase2evalFileName))

	p2, err := phase2.FromFile(phase2FileName)
	assert.NoError(t, err)

	// Check some fields of phase2 to make sure something is there
	assert.NotEmpty(t, p2.Parameters.G1.Z)
	assert.NotEmpty(t, p2.Parameters.G1.L)
	assert.NotEmpty(t, p2.Parameters.G1.Delta.X)
	assert.NotEmpty(t, p2.Parameters.G1.Delta.Y)
	assert.NotEmpty(t, p2.Parameters.G2.Delta.X)
	assert.NotEmpty(t, p2.Parameters.G2.Delta.Y)
	assert.NotEmpty(t, p2.Hash)

	eval, err := phase2.EvalFromFile(phase2evalFileName)
	assert.NoError(t, err)

	// Check some fields of phase2 evaluations to make sure something is there
	assert.NotEmpty(t, eval.G1.A)
	assert.NotEmpty(t, eval.G1.B)
	assert.NotEmpty(t, eval.G1.VKK)
	assert.NotEmpty(t, eval.G2.B)

}

func testContribute(t *testing.T) {
	phase2Contributed = make([]string, 0, 4)
	phase2Contributed = append(phase2Contributed, phase2FileName)

	for i := 0; i < 3; i++ {
		contribFileName, err := phase2.Contribute(phase2Contributed[i])
		assert.NoError(t, err)
		phase2Contributed = append(phase2Contributed, contribFileName)
	}
}

func testVerify(t *testing.T) {
	assert.NoError(t, phase2.Verify(phase2Contributed))
}

func testExtractKeys(t *testing.T) {
	lastContributionFileName := phase2Contributed[len(phase2Contributed)-1]
	assert.NoError(
		t,
		keys.Extract(
			phase1FileName, lastContributionFileName, phase2evalFileName,
			r1csFileName, pkFileName, vkFileName,
		),
	)

	pk, vk, err := keys.FromFile(pkFileName, vkFileName)
	assert.NoError(t, err)
	ccs, err := r1cs.FromFile(r1csFileName)
	assert.NoError(t, err)
	var preImage, hash fr.Element
	{
		m := native_mimc.NewMiMC()
		_, err := m.Write(preImage.Marshal())
		if err != nil {
			return
		}
		hash.SetBytes(m.Sum(nil))
	}

	witness, err := frontend.NewWitness(&TestCircuit{PreImage: preImage, Hash: hash}, ecc.BN254.ScalarField())
	assert.NoError(t, err)

	pubWitness, err := witness.Public()
	assert.NoError(t, err)

	proof, err := groth16.Prove(ccs, pk, witness)
	assert.NoError(t, err)

	err = groth16.Verify(proof, vk, pubWitness)
	assert.NoError(t, err)
}
