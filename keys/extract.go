// Package keys implements functions for interacting with proving and verification keys compatible with Gnark.
package keys

import (
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"

	"github.com/reilabs/trusted-setup/phase1"
	"github.com/reilabs/trusted-setup/phase2"
	"github.com/reilabs/trusted-setup/r1cs"
)

// Extract extracts the proving and verification keys from the given MPC objects.
//
// The input serialized Phase 1 object is given as phase1FilePath. The input serialized Phase 2 object is given as
// phase2FilePath  and it must point to the last Phase 2 object contributed to during the ceremony. The input
// serialized Phase 2 evaluations object is given as evalFilePath. The input serialized R1CS object is given
// as r1csFilePath.
//
// The output proving key is written to pkFilePath. The output verification key is written to vkFilePath.
//
// Returns nil on success and error on failure.
func Extract(phase1FilePath, phase2FilePath, evalFilePath, r1csFilePath, pkFilePath, vkFilePath string) error {
	p1, err := phase1.FromFile(phase1FilePath)
	if err != nil {
		return err
	}

	p2, err := phase2.FromFile(phase2FilePath)
	if err != nil {
		return err
	}

	eval, err := phase2.EvalFromFile(evalFilePath)
	if err != nil {
		return err
	}

	ccs, err := r1cs.FromFile(r1csFilePath)
	if err != nil {
		return err
	}

	pk, vk := mpcsetup.ExtractKeys(&p1, &p2, &eval, ccs.GetNbConstraints())
	err = ToFile(&pk, pkFilePath, &vk, vkFilePath)
	if err != nil {
		return err
	}

	return nil
}
