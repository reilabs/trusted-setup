// Package phase2 implements functions interacting with multi-party computation Phase 2 objects produced by Gnark.
package phase2

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"

	"github.com/reilabs/trusted-setup/phase1"
	"github.com/reilabs/trusted-setup/r1cs"
)

// Init initializes the multi-party computation Phase 2 object based on a serialized Phase 1 and R1CS objects.
//
// The input serialized Phase 1 object is given as phase1FilePath. The input serialized R1CS object is given as r1csFilePath.
// The output Phase 2 object is written to outputPhase2FilePath. The output Phase 2 evaluations object is written to outputEvalFilePath.
//
// Returns nil on success and error on failure.
func Init(
	phase1FilePath string, r1csFilePath string, outputPhase2FilePath string,
	outputEvalFilePath string,
) error {
	ccs, err := r1cs.FromFile(r1csFilePath)
	if err != nil {
		return err
	}

	p1, err := phase1.FromFile(phase1FilePath)
	if err != nil {
		return err
	}

	p2, eval := mpcsetup.InitPhase2(ccs.(*cs.R1CS), &p1)

	err = ToFile(p2, outputPhase2FilePath)
	if err != nil {
		return err
	}

	err = EvalToFile(eval, outputEvalFilePath)
	if err != nil {
		return err
	}

	return nil
}

// updateTimestamp appends a timestamp to the given string if not present, or updates the timestamp if present.
//
// The timestamp is appended to the string in the format: `.YYYYMMDDHHMMSS.uuuuuu`, where `u` stands for microsecond.
// For an example `foo.bar` string, the new string will be `foo.bar.202506142137.569775`, provided updateTimestamp
// is be called on June 14th, 2025 at 9:21:37.569775 PM CEST.
//
// Always succeeds. Returns the timestamped input string.
func updateTimestamp(str string) string {
	timestampRegex := regexp.MustCompile(`\.\d{14}.\d{6}$`)
	currentTimestamp := time.Now().Format("20060102150405.000000")

	if timestampRegex.MatchString(str) {
		return timestampRegex.ReplaceAllString(str, "."+currentTimestamp)
	}

	return str + "." + currentTimestamp
}

// Contribute contributes randomness to the given Phase 2 object.
//
// The Phase 2 object is deserialized from the file specified by phase2FilePath. The randomness is contributed to the
// Phase 2 object, and the updated object is written to a new file with a timestamp appended to the original file name.
// The input file is not modified.
//
// Returns the file name of the Phase 2 object containing the contributions and nil on success and empty string
// and error on failure.
func Contribute(phase2FilePath string) (newFileName string, err error) {
	p2, err := FromFile(phase2FilePath)
	if err != nil {
		return "", err
	}
	p2.Contribute()

	newFileName = updateTimestamp(phase2FilePath)
	err = ToFile(p2, newFileName)
	if err != nil {
		return "", err
	}
	return newFileName, err
}

// Verify verifies the given Phase 2 objects for the correctness of their contributions.
//
// The Phase 2 objects are deserialized from the files specified by phase2FilePaths. The verification is performed
// between each consecutive pair of Phase 2 objects. The verification is performed in the order of the input files.
//
// Returns nil on success and error on failure.
func Verify(phase2FilePaths []string) error {
	if len(phase2FilePaths) < 2 {
		return fmt.Errorf("at least two phase 2 files must be provided")
	}

	phase2s := make([]*mpcsetup.Phase2, 0, len(phase2FilePaths))
	for _, phase2FilePath := range phase2FilePaths {
		p2, err := FromFile(phase2FilePath)
		if err != nil {
			return err
		}
		phase2s = append(phase2s, &p2)
	}
	for i := 0; i < len(phase2s)-1; i++ {
		logMsg := fmt.Sprintf("Verifying: %s, %s...", phase2FilePaths[i], phase2FilePaths[i+1])
		err := mpcsetup.VerifyPhase2(phase2s[i], phase2s[i+1])
		if err != nil {
			log.Println(logMsg, "failed")
			return err
		}
		log.Println(logMsg, "ok")
	}

	return nil
}
