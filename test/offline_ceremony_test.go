package test

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/phase1"
	"github.com/reilabs/trusted-setup/phase2"
	"github.com/reilabs/trusted-setup/r1cs"
)

// testOfflineCeremony verifies the trusted setup ceremony in the offline mode.
//
// The whole ceremony is done synchronously, step by step, from converting ptau to Phase 1,
// through Phase 2 initialization, contribution and verification. In the end, keys are extracted
// proof is created and verified. Every step is explicitly called, as if commands were issued
// manually by the Coordinator and Contributors.
func TestOfflineCeremony(t *testing.T) {
	setup()

	t.Run("Ptau", testPtau)
	t.Run("Init", testInit)
	t.Run("Contribute", testContribute)
	t.Run("Verify", testVerify)
	t.Run("Extract keys", testExtractKeys)
	t.Run("Prove and verify", testProveAndVerify)

	teardown()
}

const phase1FileName = "test.phase1"
const phase2FileName = "test.phase2"
const srsCommonsFileName = "test.srscommons"
const r1csFileName = "test.r1cs"
const pkFileName = "test.pk"
const vkFileName = "test.vk"

func setup() {
	ccs, err := buildCcs()
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
		srsCommonsFileName,
		r1csFileName,
		pkFileName,
		vkFileName,
	}

	matches, err := filepath.Glob(phase2FileName + "*")
	if err == nil {
		filesToRemove = append(filesToRemove, matches...)
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
	const ptauFileName = "resources/test.ptau"

	assert.NoError(t, phase1.FromPtau(ptauFileName, phase1FileName))

	p1, err := phase1.FromFile(phase1FileName)
	assert.NoError(t, err)

	// Phase1 has no public fields, so let's assume it is correct if we can
	// successfully contribute to it.
	p1beforeContribution, err := phase1.FromFile(phase1FileName)
	assert.NoError(t, err)
	p1.Contribute()
	err = p1beforeContribution.Verify(&p1)
	assert.NoError(t, err)
}

func testInit(t *testing.T) {
	assert.NoError(t, phase2.Init(phase1FileName, r1csFileName, phase2FileName, srsCommonsFileName))

	p2, err := phase2.FromFile(phase2FileName)
	assert.NoError(t, err)

	// Check some fields of phase2 to make sure something is there
	assert.NotEmpty(t, p2.Parameters.G1.Z)
	assert.NotEmpty(t, p2.Parameters.G1.PKK)
	assert.NotEmpty(t, p2.Parameters.G1.Delta.X)
	assert.NotEmpty(t, p2.Parameters.G1.Delta.Y)
	assert.NotEmpty(t, p2.Parameters.G2.Delta.X)
	assert.NotEmpty(t, p2.Parameters.G2.Delta.Y)

	srsCommons, err := phase2.SrsCommonsFromFile(srsCommonsFileName)
	assert.NoError(t, err)

	// Check some fields of SRS commons to make sure something is there
	assert.NotEmpty(t, srsCommons.G1.Tau)
	assert.NotEmpty(t, srsCommons.G1.BetaTau)
	assert.NotEmpty(t, srsCommons.G1.AlphaTau)
	assert.NotEmpty(t, srsCommons.G2.Beta)
	assert.NotEmpty(t, srsCommons.G2.Tau)
}

func testContribute(t *testing.T) {
	phase2Contributions := []string{phase2FileName}

	for i := 0; i < 3; i++ {
		contribFileName := strings.Join([]string{phase2FileName, strconv.Itoa(i)}, ".")
		assert.NoError(t, phase2.Contribute(phase2Contributions[i], contribFileName))
		phase2Contributions = append(phase2Contributions, contribFileName)
	}
}

func testVerify(t *testing.T) {
	phase2Contributions := []string{phase2FileName}
	for i := 0; i < 3; i++ {
		contribFileName := strings.Join([]string{phase2FileName, strconv.Itoa(i)}, ".")
		phase2Contributions = append(phase2Contributions, contribFileName)
	}

	for i := 1; i < 3; i++ {
		err := phase2.Verify(phase2Contributions[i], phase2Contributions[i+1])
		assert.NoError(t, err)
	}
}

func testExtractKeys(t *testing.T) {
	phase2Contributions := []string{phase2FileName}
	for i := 0; i < 3; i++ {
		contribFileName := strings.Join([]string{phase2FileName, strconv.Itoa(i)}, ".")
		phase2Contributions = append(phase2Contributions, contribFileName)
	}

	assert.NoError(
		t,
		phase2.ExtractKeys(
			r1csFileName, srsCommonsFileName, phase2Contributions[1:], pkFileName, vkFileName,
		),
	)
}

func testProveAndVerify(t *testing.T) {
	pk, vk, err := phase2.PkVkFromFile(pkFileName, vkFileName)
	assert.NoError(t, err)
	ccs, err := r1cs.FromFile(r1csFileName)
	assert.NoError(t, err)

	err = proveAndVerify(ccs, pk, vk)
	assert.NoError(t, err)
}
