package test_circuit

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	native_mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/groth16"
	cs "github.com/consensys/gnark/constraint/bn254"
	"github.com/consensys/gnark/frontend"
	gnark_r1cs "github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/hash/mimc"
)

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

func BuildCcs() (*cs.R1CS, error) {
	circuit := &TestCircuit{}
	r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), gnark_r1cs.NewBuilder, circuit)
	return r1cs.(*cs.R1CS), err
}

func ProveAndVerify(ccs *cs.R1CS, pk *groth16.ProvingKey, vk *groth16.VerifyingKey) error {
	var preImage, hash fr.Element
	m := native_mimc.NewMiMC()
	_, err := m.Write(preImage.Marshal())
	if err != nil {
		return err
	}
	hash.SetBytes(m.Sum(nil))

	witness, err := frontend.NewWitness(&TestCircuit{PreImage: preImage, Hash: hash}, ecc.BN254.ScalarField())
	if err != nil {
		return err
	}

	pubWitness, err := witness.Public()
	if err != nil {
		return err
	}

	proof, err := groth16.Prove(ccs, *pk, witness)
	if err != nil {
		return err
	}

	err = groth16.Verify(proof, *vk, pubWitness)
	if err != nil {
		return err
	}

	return nil
}
