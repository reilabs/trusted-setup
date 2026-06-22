// import_r1cs takes a portable JSON R1CS dump (typically produced by the
// Provekit `export-gnark-r1cs` CLI command) and reconstructs it into gnark's
// `cs.R1CS` constraint system on BN254, then writes the result as a
// gnark-binary `.r1cs` file ready to be fed to `./trusted-setup init`.
//
// This is the Phase-2-precondition bridge: gnark's ceremony needs an R1CS in
// its own format; Provekit produces R1CS in arkworks format. The Rust side
// dumps the canonical (A, B, C, public-count, commitments) data; this tool
// rebuilds gnark's struct using gnark's public API and serializes via
// gnark's own WriteTo so the file is byte-identical to what gnark would emit
// for the same circuit if it had been compiled inside gnark.
//
// Usage:
//
//	./import_r1cs --in my_circuit.r1cs.json --out my_circuit.r1cs
//
// Then:
//
//	./trusted-setup init --phase1 prod.ph1 --r1cs my_circuit.r1cs \
//	                     --phase2 ceremony.ph2 --srscommons ceremony.srs
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/constraint"
	cs "github.com/consensys/gnark/constraint/bn254"
)

// -----------------------------------------------------------------------------
// JSON schema
// -----------------------------------------------------------------------------

// dump is the portable intermediate the Rust side emits. Field names match a
// snake_case convention so the Rust serde_json layer maps without renaming.
type dump struct {
	NumPublicInputs int         `json:"num_public_inputs"`
	NumWitnesses    int         `json:"num_witnesses"`
	NumConstraints  int         `json:"num_constraints"`
	Constraints     []rawR1C    `json:"constraints"`
	Commitments     []rawCommit `json:"commitments,omitempty"`
}

// rawR1C is one constraint row: A · B == C, each side a sparse linear
// combination of (wire_id, coefficient_hex).
type rawR1C struct {
	A []rawTerm `json:"a"`
	B []rawTerm `json:"b"`
	C []rawTerm `json:"c"`
}

// rawTerm uses a 2-element tuple [wire, coeff_hex] to keep the JSON dump
// compact for large circuits — millions of constraints turn into JSON files
// fast and object-per-term doubles the byte count for no reader benefit.
type rawTerm [2]json.RawMessage

func (t *rawTerm) parse() (wire uint32, coeffHex string, err error) {
	var w int
	if err = json.Unmarshal(t[0], &w); err != nil {
		return
	}
	if w < 0 {
		err = fmt.Errorf("negative wire id %d", w)
		return
	}
	if err = json.Unmarshal(t[1], &coeffHex); err != nil {
		return
	}
	return uint32(w), coeffHex, nil
}

// rawCommit mirrors gnark's `constraint.Groth16Commitment`. Provekit emits
// one entry per gnark commitment (which, for the N-challenges-from-1-real-
// commitment circuit pattern, is N entries — the first carries the full
// PrivateCommitted list, the rest are dummies with empty PrivateCommitted
// whose only purpose is to allocate the corresponding K-base in vk.G1.K).
// The N-challenge derivation that Provekit does over its single real
// commitment is invisible to gnark and happens at prove time inside
// Provekit's prover.
type rawCommit struct {
	PublicAndCommitmentCommitted []int `json:"public_and_commitment_committed"`
	PrivateCommitted             []int `json:"private_committed"`
	NbPublicCommitted            int   `json:"nb_public_committed"`
	// CommitmentWire is the wire-id gnark treats as holding the BSB22
	// commitment value (`Groth16Commitment.CommitmentIndex`). Provekit picks
	// the i-th sorted challenge wire for the i-th commitment entry.
	CommitmentWire int `json:"commitment_wire"`
}

// -----------------------------------------------------------------------------
// main
// -----------------------------------------------------------------------------

func main() {
	in := flag.String("in", "", "input JSON R1CS dump")
	out := flag.String("out", "", "output gnark .r1cs binary")
	flag.Parse()
	if *in == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: import_r1cs --in <dump.json> --out <file.r1cs>")
		os.Exit(2)
	}

	d, err := loadDump(*in)
	if err != nil {
		log.Fatalf("read %s: %v", *in, err)
	}

	r1cs, err := buildR1CS(d)
	if err != nil {
		log.Fatalf("build r1cs: %v", err)
	}

	if err := writeR1CS(r1cs, *out); err != nil {
		log.Fatalf("write %s: %v", *out, err)
	}

	log.Printf("wrote %s", *out)
	log.Printf("  constraints:        %d", r1cs.GetNbConstraints())
	log.Printf("  public variables:   %d (includes ONE_WIRE)", r1cs.GetNbPublicVariables())
	log.Printf("  secret variables:   %d", r1cs.GetNbSecretVariables())
	log.Printf("  internal variables: %d", r1cs.GetNbInternalVariables())
	log.Printf("  total wires:        %d",
		r1cs.GetNbPublicVariables()+r1cs.GetNbSecretVariables()+r1cs.GetNbInternalVariables())
	log.Printf("  commitments:        %d", len(d.Commitments))
}

func loadDump(path string) (*dump, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var d dump
	if err := json.NewDecoder(f).Decode(&d); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	if d.NumConstraints != len(d.Constraints) {
		return nil, fmt.Errorf(
			"num_constraints (%d) != len(constraints) (%d)",
			d.NumConstraints, len(d.Constraints))
	}
	if d.NumWitnesses < 1+d.NumPublicInputs {
		return nil, fmt.Errorf(
			"num_witnesses (%d) < 1 + num_public_inputs (%d) — missing ONE_WIRE",
			d.NumWitnesses, d.NumPublicInputs)
	}
	return &d, nil
}

// -----------------------------------------------------------------------------
// R1CS construction
// -----------------------------------------------------------------------------

// buildR1CS reconstructs gnark's R1CS from the JSON dump.
//
// Wire layout produced here matches gnark's frontend convention:
//
//	[0]                                            ONE_WIRE
//	[1 ..= num_public_inputs]                      user public inputs
//	[num_public_inputs+1 ..]                       secret wires
//
// All non-public wires are classified as **secret** rather than internal.
// This is a deliberate choice: gnark's `BlueprintGenericR1C` updates an
// "instruction tree" that tracks each internal wire as being produced by
// one specific constraint. An R1CS where the same wire appears on multiple
// sides of one constraint (e.g. `x*x = y` puts wire `x` in both L and R)
// makes that tracker panic. Secret wires are treated as system inputs and
// the tracker skips them, so the same R1CS imports cleanly. For ceremony
// (Phase 2) purposes the secret/internal split is irrelevant — gnark
// groups them together when sizing `pk.G1.K`. The trade-off: an R1CS
// built this way cannot be used directly by `groth16.Prove` because the
// witness assignment would expect secret values for every non-public wire,
// not just the user-private ones. For our use case (run the ceremony,
// then load the keys into Provekit) that's fine.
func buildR1CS(d *dump) (*cs.R1CS, error) {
	system := cs.NewR1CS(d.NumConstraints)

	// 1. Public variables. AddPublicVariable returns sequential wire IDs
	//    starting at 0; the first one is the ONE_WIRE.
	system.AddPublicVariable("1")
	for i := 0; i < d.NumPublicInputs; i++ {
		system.AddPublicVariable(fmt.Sprintf("pub_%d", i))
	}

	// 2. All remaining wires as secret. See function docstring for why.
	nbNonPublic := d.NumWitnesses - (1 + d.NumPublicInputs)
	for i := 0; i < nbNonPublic; i++ {
		_ = system.AddSecretVariable(fmt.Sprintf("sec_%d", i))
	}

	// 3. Register the generic R1C blueprint. Every AddR1C call routes
	//    through this blueprint; we keep one ID for the whole circuit
	//    rather than a fresh blueprint per constraint.
	r1cBpID := system.AddBlueprint(&constraint.BlueprintGenericR1C{})

	// 4. Coefficient interning. gnark's coefficient table deduplicates
	//    repeated coefficients (an R1CS typically has many fewer distinct
	//    values than terms); the table maps coeff bytes → uint32 CID.
	//    We re-intern locally to avoid re-parsing the same hex over and
	//    over.
	coeffCache := make(map[string]uint32, 64)
	addCoeff := func(hexStr string) (uint32, error) {
		if cid, ok := coeffCache[hexStr]; ok {
			return cid, nil
		}
		fe, err := parseFrHex(hexStr)
		if err != nil {
			return 0, err
		}
		// gnark's generic constraint system stores coefficients as
		// `constraint.U64` ([6]uint64) so it can host larger fields
		// uniformly. For BN254 only the first 4 limbs are populated;
		// `CoeffTable.AddCoeff` re-views them as `fr.Element` via a slice
		// cast (constraint/bn254/coeff.go:78).
		var u64 constraint.U64
		copy(u64[:], fe[:])
		cid := system.AddCoeff(u64)
		coeffCache[hexStr] = cid
		return cid, nil
	}

	totalWires := uint32(d.NumWitnesses)
	buildLinExp := func(terms []rawTerm, label string, ci int) (constraint.LinearExpression, error) {
		out := make(constraint.LinearExpression, len(terms))
		for ti := range terms {
			wire, coeffHex, err := terms[ti].parse()
			if err != nil {
				return nil, fmt.Errorf("constraint[%d].%s[%d]: %w", ci, label, ti, err)
			}
			if wire >= totalWires {
				return nil, fmt.Errorf(
					"constraint[%d].%s[%d]: wire id %d >= num_witnesses %d",
					ci, label, ti, wire, totalWires)
			}
			cid, err := addCoeff(coeffHex)
			if err != nil {
				return nil, fmt.Errorf(
					"constraint[%d].%s[%d]: coeff %q: %w",
					ci, label, ti, coeffHex, err)
			}
			out[ti] = constraint.Term{CID: cid, VID: wire}
		}
		return out, nil
	}

	// 5. Add constraints in order.
	for ci, c := range d.Constraints {
		la, err := buildLinExp(c.A, "a", ci)
		if err != nil {
			return nil, err
		}
		lb, err := buildLinExp(c.B, "b", ci)
		if err != nil {
			return nil, err
		}
		lc, err := buildLinExp(c.C, "c", ci)
		if err != nil {
			return nil, err
		}
		_ = system.AddR1C(constraint.R1C{L: la, R: lb, O: lc}, r1cBpID)
	}

	// 6. BSB22 commitments. Each JSON entry is registered with gnark as a
	//    `constraint.Groth16Commitment`. Phase 2 then allocates one
	//    Pedersen basis (size = len(PrivateCommitted)) and appends one
	//    entry to vk.G1.K per registration. Provekit's N-challenges-from-
	//    1-real-commitment circuit pattern expects the Rust exporter to
	//    emit N entries: entry [0] carries the full PrivateCommitted list,
	//    entries [1..N] have empty PrivateCommitted (dummy commitments
	//    whose only purpose is the extra K-base slot). All N entries must
	//    be in ascending CommitmentIndex order — gnark's wire-iteration
	//    matcher at gnark/backend/groth16/bn254/mpcsetup/phase2.go:289
	//    advances `nbCommitmentsSeen` only when the next commitment's
	//    CommitmentIndex matches the current wire index, so out-of-order
	//    entries get silently skipped.
	for ci, c := range d.Commitments {
		if c.NbPublicCommitted < 0 || c.NbPublicCommitted > len(c.PublicAndCommitmentCommitted) {
			return nil, fmt.Errorf(
				"commitment[%d]: nb_public_committed (%d) out of range for "+
					"public_and_commitment_committed (len %d)",
				ci, c.NbPublicCommitted, len(c.PublicAndCommitmentCommitted))
		}
		if c.CommitmentWire < 0 || uint32(c.CommitmentWire) >= totalWires {
			return nil, fmt.Errorf(
				"commitment[%d]: commitment_wire %d outside [0, %d)",
				ci, c.CommitmentWire, totalWires)
		}
		if ci > 0 && c.CommitmentWire <= d.Commitments[ci-1].CommitmentWire {
			return nil, fmt.Errorf(
				"commitment[%d].commitment_wire (%d) must be > commitment[%d].commitment_wire (%d) "+
					"— gnark requires ascending CommitmentIndex across registered commitments",
				ci, c.CommitmentWire, ci-1, d.Commitments[ci-1].CommitmentWire)
		}
		// PrivateCommitted must be strictly ascending: gnark's Phase 2 feeds
		// it to internal.NewMergeIterator, which "assumes that all slices ...
		// are sorted" (gnark backend/groth16/internal/utils.go) and only ever
		// peeks each slice's head. An unsorted or duplicate entry is neither
		// rejected nor panicked on — the wire-walk silently misplaces the
		// committed contribution, yielding a wrong proving/verifying key. We
		// verify here rather than trust the exporter.
		for i, w := range c.PrivateCommitted {
			if w < 0 || uint32(w) >= totalWires {
				return nil, fmt.Errorf(
					"commitment[%d]: private_committed wire %d outside [0, %d)",
					ci, w, totalWires)
			}
			if i > 0 && w <= c.PrivateCommitted[i-1] {
				return nil, fmt.Errorf(
					"commitment[%d]: private_committed must be strictly ascending, "+
						"but wire %d at index %d is <= previous wire %d "+
						"(gnark's MergeIterator assumes sorted, unique entries)",
					ci, w, i, c.PrivateCommitted[i-1])
			}
		}
		for _, w := range c.PublicAndCommitmentCommitted {
			if w < 0 || uint32(w) >= totalWires {
				return nil, fmt.Errorf(
					"commitment[%d]: public_and_commitment_committed wire %d "+
						"outside [0, %d)",
					ci, w, totalWires)
			}
		}
		err := system.AddCommitment(constraint.Groth16Commitment{
			PublicAndCommitmentCommitted: c.PublicAndCommitmentCommitted,
			PrivateCommitted:             c.PrivateCommitted,
			NbPublicCommitted:            c.NbPublicCommitted,
			CommitmentIndex:              c.CommitmentWire,
		})
		if err != nil {
			return nil, fmt.Errorf("commitment[%d]: AddCommitment: %w", ci, err)
		}
	}

	// Sanity check: gnark counted the same number of constraints we added.
	if got := system.GetNbConstraints(); got != d.NumConstraints {
		return nil, fmt.Errorf(
			"constraint count mismatch after build: gnark sees %d, dump claimed %d",
			got, d.NumConstraints)
	}

	return system, nil
}

// -----------------------------------------------------------------------------
// Fr parsing
// -----------------------------------------------------------------------------

// parseFrHex accepts "0x"-prefixed (or bare) big-endian hex and turns it into
// a canonical fr.Element. Rejects values >= the BN254 scalar modulus.
func parseFrHex(s string) (fr.Element, error) {
	var fe fr.Element
	hexStr := strings.TrimPrefix(s, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")
	if len(hexStr)%2 == 1 {
		hexStr = "0" + hexStr
	}
	if hexStr == "" {
		hexStr = "00"
	}
	bs, err := hex.DecodeString(hexStr)
	if err != nil {
		return fe, fmt.Errorf("decode hex %q: %w", s, err)
	}
	// fr.Element.SetBytes accepts big-endian and reduces mod r. To enforce
	// canonical input we compare round-trip bytes — anything that came in
	// non-canonical (>= r) would be silently reduced and the comparison
	// catches that.
	if len(bs) > fr.Bytes {
		// strip leading zero padding past 32 bytes; anything non-zero past
		// the field width is a hard error
		i := 0
		for i < len(bs)-fr.Bytes && bs[i] == 0 {
			i++
		}
		if len(bs)-i > fr.Bytes {
			return fe, fmt.Errorf("coefficient %q exceeds 32 bytes after stripping zeros", s)
		}
		bs = bs[i:]
	}
	fe.SetBytes(bs)

	// Verify canonical: re-serialize and zero-pad input to 32 bytes, then
	// compare. Catches the "input was >= modulus" case.
	padded := make([]byte, fr.Bytes)
	copy(padded[fr.Bytes-len(bs):], bs)
	out := fe.Bytes()
	for i := 0; i < fr.Bytes; i++ {
		if out[i] != padded[i] {
			return fe, fmt.Errorf("non-canonical coefficient %q (>= modulus)", s)
		}
	}
	return fe, nil
}

// -----------------------------------------------------------------------------
// Write
// -----------------------------------------------------------------------------

// writeR1CS serializes via gnark's own WriteTo so the file format matches
// exactly what the trusted-setup tool expects from its `--r1cs` flag.
func writeR1CS(r1cs *cs.R1CS, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeAll(f, r1cs)
}

func writeAll(w io.Writer, r1cs *cs.R1CS) error {
	if _, err := r1cs.WriteTo(w); err != nil {
		return fmt.Errorf("gnark WriteTo: %w", err)
	}
	return nil
}
