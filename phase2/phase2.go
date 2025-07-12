// Package phase2 implements functions interacting with multi-party computation Phase 2 objects produced by Gnark.
package phase2

import (
	"fmt"
	"log"

	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"

	"github.com/reilabs/trusted-setup/phase1"
	"github.com/reilabs/trusted-setup/r1cs"
	"github.com/reilabs/trusted-setup/randomness"
)

// Init initializes the multi-party computation Phase 2 object based on a serialized Phase 1 and R1CS objects.
//
// The input serialized Phase 1 object is given as phase1FilePath. The input serialized R1CS object is given as r1csFilePath.
// The output Phase 2 object is written to outputPhase2FilePath.
//
// Returns nil on success and error on failure.
func Init(
	phase1FilePath, r1csFilePath, outputPhase2FilePath, outputSrsCommonsPath string,
) error {
	ccs, err := r1cs.FromFile(r1csFilePath)
	if err != nil {
		return err
	}

	p1, err := phase1.FromFile(phase1FilePath)
	if err != nil {
		return err
	}

	log.Printf("Generating SRS commons form Phase 1 (beacon: %x)", randomness.GetBeacon())
	srsCommons := p1.Seal(randomness.GetBeacon())
	err = SrsCommonsToFile(srsCommons, outputSrsCommonsPath)
	if err != nil {
		return err
	}

	log.Print("Initializing Phase 2")
	p2 := mpcsetup.Phase2{}
	_ = p2.Initialize(ccs.(*cs.R1CS), &srsCommons)
	err = ToFile(p2, outputPhase2FilePath)
	if err != nil {
		return err
	}

	return nil
}

// Contribute contributes randomness to the given Phase 2 object.
//
// The Phase 2 object is deserialized from the file specified by inputPhase2FilePath. The randomness is contributed to
// the Phase 2 object, and the updated object is written outputPhase2FilePath. The input file is not modified.
func Contribute(inputPhase2FilePath, outputPhase2FilePath string) error {
	p2, err := FromFile(inputPhase2FilePath)
	if err != nil {
		return err
	}
	log.Print("Contributing randomness to Phase 2")
	p2.Contribute()

	err = ToFile(p2, outputPhase2FilePath)
	if err != nil {
		return err
	}
	return err
}

// Verify verifies the given Phase 2 objects for the correctness of their contributions.
//
// The Phase 2 objects are deserialized from the files specified by phase2prevFilePath and phase2nextFilePath.
// phase2prevFilePath is a file that was an input to a contribution, and phase2nextFilePath is the output of that
// contribution.
//
// Returns nil on success and error on failure.
func Verify(phase2prevFilePath, phase2nextFilePath string) error {
	prev, err := FromFile(phase2prevFilePath)
	if err != nil {
		return err
	}
	next, err := FromFile(phase2nextFilePath)
	if err != nil {
		return err
	}

	log.Print("Verifying the most recent Phase 2 against the previous step")
	err = prev.Verify(&next)
	if err != nil {
		return err
	}

	return nil
}

// ExtractKeys verifies the given Phase 2 objects for the correctness of their contributions and extracts the proving
// and verification keys.
//
// The Phase 2 objects are deserialized from the files specified by phase2FilePaths. The verification is performed
// between each consecutive pair of Phase 2 objects that contain contributions (that is EXCLUDING the initial one).
// The verification is performed in the order of the input files.
//
// The constraint system used for Phase 2 initialization and the SRS Commons object being the result if the initialization
// must be provided in the form of file paths.
//
// The output proving key is written to outputPkFilePath. The output verification key is written to outputVkFilePath.
//
// Returns nil on success and error on failure.
func ExtractKeys(
	r1csFilePath, srsCommonsFilePath string, phase2FilePaths []string, outputPkFilePath, outputVkFilePath string,
) error {
	if len(phase2FilePaths) < 2 {
		return fmt.Errorf("at least two phase 2 files must be provided")
	}
	ccs, err := r1cs.FromFile(r1csFilePath)
	if err != nil {
		return err
	}

	srsCommons, err := SrsCommonsFromFile(srsCommonsFilePath)
	if err != nil {
		return err
	}

	phase2s := make([]*mpcsetup.Phase2, 0, len(phase2FilePaths))
	for _, phase2FilePath := range phase2FilePaths {
		p2, err := FromFile(phase2FilePath)
		if err != nil {
			return err
		}
		phase2s = append(phase2s, &p2)
	}

	log.Printf("Verifying all Phase 2 contributions and generating Keys (beacon: %x)", randomness.GetBeacon())
	pk, vk, err := mpcsetup.VerifyPhase2(ccs.(*cs.R1CS), &srsCommons, randomness.GetBeacon(), phase2s...)
	if err != nil {
		return err
	}

	err = PkVkToFile(pk, outputPkFilePath, vk, outputVkFilePath)
	if err != nil {
		return err
	}

	return nil
}
