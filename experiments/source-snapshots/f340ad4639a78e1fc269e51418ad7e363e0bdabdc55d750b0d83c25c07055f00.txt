// Package cryptoref is a bounded in-memory correctness reference for permuting
// codewords and masking binary-field values. It reuses the prototype AES helpers;
// it is not a reviewed cryptographic suite or a large-database backend.
//
// Only the behavior (storage-position permutation and separate XOR masks) is a
// reference for future implementations. Streaming, external-memory processing,
// storage layout, and production APIs need not follow this implementation.
package cryptoref

import (
	"crypto/rand"
	"fmt"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/utils"
)

// MaxLength bounds the in-memory permutation and stored codeword fixtures.
const MaxLength = 256

// GenerateKey returns a fresh 32-byte AES key from crypto/rand. Call separately
// for each query-sharing group's permutation and for every PIR database encoding's
// masks. Re-encoding changed data requires a new masking key, even at unchanged
// coordinates. Keys are secret material, not database identifiers.
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate reference AES key: %w", err)
	}
	return key, nil
}

// Permutation is an immutable codeword-position -> storage-position map.
// It contains secret information and must not be included in public metadata.
// Obtain one through NewPermutation; its zero value is not initialized.
type Permutation struct {
	forward []int
}

// NewPermutation snapshots the existing AES-based prototype permutation with
// an empty tweak over [0,length). It verifies bijectivity on the entire tiny
// domain. The map, not the mutable helper, is retained, so Map is concurrent-safe.
// length must be in [2,MaxLength]; key must contain 16, 24, or 32 bytes.
// This check establishes a permutation, not its cryptographic strength.
func NewPermutation(length int, key []byte) (*Permutation, error) {
	if err := checkLength(length); err != nil {
		return nil, err
	}
	helper, err := utils.NewPermutator(length, key, nil)
	if err != nil {
		return nil, fmt.Errorf("create reference permutation: %w", err)
	}
	p := &Permutation{forward: make([]int, length)}
	seen := make([]bool, length)
	for i := range p.forward {
		position, err := helper.Permute(uint64(i))
		if err != nil {
			return nil, fmt.Errorf("permute position %d: %w", i, err)
		}
		if position >= uint64(length) || seen[int(position)] {
			return nil, fmt.Errorf("prototype map is not a permutation at position %d", i)
		}
		p.forward[i] = int(position)
		seen[int(position)] = true
	}
	return p, nil
}

// Map returns the storage positions for codeword positions, in the same order
// and with repetitions retained. It does not modify or alias the input.
func (p *Permutation) Map(positions []int) ([]int, error) {
	if p == nil || len(p.forward) == 0 {
		return nil, fmt.Errorf("uninitialized reference permutation")
	}
	for _, position := range positions {
		if position < 0 || position >= len(p.forward) {
			return nil, fmt.Errorf("codeword position %d is outside [0,%d)", position, len(p.forward))
		}
	}
	out := make([]int, len(positions))
	for i, position := range positions {
		out[i] = p.forward[position]
	}
	return out, nil
}

// Masker holds the secret AES mask state for one PIR database encoding.
// A query-sharing group can use the same Permutation, but each database needs a
// different masking key. Never use this Masker/key for changed data or another
// database. Constructors cannot detect key reuse across independent calls.
// Encode and Unmask are safe for concurrent use with caller-owned slices.
type Masker struct {
	mask   utils.Encryptor
	length int
	q      uint32
}

// NewMasker accepts fixture field degrees 2, 3, or 4 and lengths [2,MaxLength].
// The key is copied into the cipher state and must contain 16, 24, or 32 bytes.
// GenerateKey provides fresh keys; fixed keys are useful for explicit test vectors.
func NewMasker(fieldDegree, length int, key []byte) (*Masker, error) {
	if fieldDegree < 2 || fieldDegree > 4 {
		return nil, fmt.Errorf("reference field degree must be 2, 3, or 4, got %d", fieldDegree)
	}
	if err := checkLength(length); err != nil {
		return nil, err
	}
	q := uint32(1) << uint(fieldDegree)
	mask, err := utils.NewEncryptor(key, int(q))
	if err != nil {
		return nil, fmt.Errorf("create reference mask cipher: %w", err)
	}
	return &Masker{mask: mask, length: length, q: q}, nil
}

// Encode returns a separate stored codeword with
// stored[permutation(i)] = codeword[i] XOR mask(permutation(i)).
// Values must be field-element integers in [0,q). No application input conversion,
// code encoding, file format, or storage access occurs here. Repeating this call
// is only valid for the same database data under the same masking key.
func (m *Masker) Encode(codeword []uint32, permutation *Permutation) ([]uint32, error) {
	if err := m.initialized(); err != nil {
		return nil, err
	}
	if len(codeword) != m.length {
		return nil, fmt.Errorf("codeword has %d elements, want %d", len(codeword), m.length)
	}
	if permutation == nil || len(permutation.forward) != m.length {
		return nil, fmt.Errorf("permutation length must match codeword length %d", m.length)
	}
	for i, value := range codeword {
		if value >= m.q {
			return nil, fmt.Errorf("codeword element %d is outside [0,%d)", i, m.q)
		}
	}
	stored := make([]uint32, m.length)
	for i, value := range codeword {
		position := permutation.forward[i]
		stored[position] = value ^ m.maskAt(position)
	}
	return stored, nil
}

// Unmask removes masks from ordinary storage read responses. positions contains
// STORAGE positions, not codeword positions. Response order and repetitions are
// preserved; the client still performs code decoding. This does not authenticate
// contents: in-range corruption cannot be detected by masking alone.
func (m *Masker) Unmask(positions []int, values []uint32) ([]uint32, error) {
	if err := m.initialized(); err != nil {
		return nil, err
	}
	if len(positions) != len(values) {
		return nil, fmt.Errorf("received %d values for %d positions", len(values), len(positions))
	}
	for i, position := range positions {
		if position < 0 || position >= m.length {
			return nil, fmt.Errorf("storage position %d is outside [0,%d)", position, m.length)
		}
		if values[i] >= m.q {
			return nil, fmt.Errorf("response element %d is outside [0,%d)", i, m.q)
		}
	}
	out := make([]uint32, len(values))
	for i, position := range positions {
		out[i] = values[i] ^ m.maskAt(position)
	}
	return out, nil
}

// Keep the prototype PRF's precise format: AES_k(LE64(storage position)||LE64(0)),
// interpreted big-endian and reduced modulo q. Since q is a power of two, this
// selects the low fieldDegree bits without bias. The SECOND operation, adding
// that field mask to data, must be XOR rather than integer addition modulo q.
// Separate database/encoding keys supply the domain separation; the fixed zero
// here is not a shared database identifier.
func (m *Masker) maskAt(position int) uint32 {
	return uint32(m.mask.EncryptPosition(0, position))
}

func (m *Masker) initialized() error {
	if m == nil || m.length == 0 {
		return fmt.Errorf("uninitialized reference masker")
	}
	return nil
}

func checkLength(length int) error {
	if length < 2 || length > MaxLength {
		return fmt.Errorf("reference length must be in [2,%d], got %d", MaxLength, length)
	}
	return nil
}
