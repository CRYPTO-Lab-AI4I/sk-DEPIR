package liftrs

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// LoadScalingOnline reads the experimental LRSprobe1 minimal plan with bounded
// block decoding and a mandatory expected SHA-256 from a trusted manifest.
// Checksums establish byte identity, not mathematical validity. Only load files
// derived from verified encoders. This is not a stable public artifact format.
// maxElements limits N before allocation; input must end exactly after spectrum.
// K must permit the ordinary prefix convolution to fit one field domain (2K-1≤N).
func LoadScalingOnline(source io.Reader, digest string, maxElements int) (*ScalingOnlinePlan, error) {
	hash := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(source, hash), 64<<10)
	var magic [9]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if string(magic[:]) != "LRSprobe1" {
		return nil, fmt.Errorf("bad probe magic")
	}
	var header [4]uint32
	for i := range header {
		if err := binary.Read(r, binary.LittleEndian, &header[i]); err != nil {
			return nil, err
		}
	}
	s, t, k, n := int(header[0]), int(header[1]), int(header[2]), int(header[3])
	if s < 2 || s > 14 || t < 1 || t >= (1<<uint(s))-1 || n != 1<<uint(2*s) || n > maxElements || k < 1 || k > n/2 {
		return nil, fmt.Errorf("invalid online probe shape/limit")
	}
	polynomial, _ := BinaryPolynomial(s)
	base, extension, delta, omega := orbitFields(s, polynomial)
	b := OrbitBasis{spec: Spec{s, 2, t}, polynomial: polynomial, delta: delta, omega: omega, base: base, extension: extension, extensionFFT: newSparseTransform(extension)}
	p := &ScalingOnlinePlan{basis: b, k: k, p: make([]uint32, k+1), spectrum: make([]uint32, n)}
	for j, values := range [][]uint32{p.p, p.spectrum} {
		limit := base.q
		if j == 1 {
			limit = extension.q
		}
		var block [65536]byte
		for start := 0; start < len(values); start += len(block) / 4 {
			count := minInt(len(block)/4, len(values)-start)
			if _, err := io.ReadFull(r, block[:4*count]); err != nil {
				return nil, err
			}
			for j := 0; j < count; j++ {
				x := binary.LittleEndian.Uint32(block[4*j:])
				if x >= limit {
					return nil, fmt.Errorf("online element outside field")
				}
				values[start+j] = x
			}
		}
	}
	if p.p[0] != 1 {
		return nil, fmt.Errorf("invalid P constant")
	}
	if _, err := r.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("trailing probe data or read error")
	}
	if hex.EncodeToString(hash.Sum(nil)) != digest {
		return nil, fmt.Errorf("online probe checksum differs from manifest")
	}
	return p, nil
}

// Dimension returns the number of field elements supplied to Encode.
func (p *ScalingOnlinePlan) Dimension() int { return p.k }

// Length returns the canonical codeword's field-element count.
func (p *ScalingOnlinePlan) Length() int { return p.basis.Length() }

// FieldDegree returns s for GF(2^s) input/output elements.
func (p *ScalingOnlinePlan) FieldDegree() int { return p.basis.spec.FieldDegree }

// SystematicPosition maps a zero-based input index to its canonical codeword coordinate.
func (p *ScalingOnlinePlan) SystematicPosition(index int) (uint64, error) {
	if index < 0 || index >= p.k {
		return 0, fmt.Errorf("systematic index outside message")
	}
	return uint64(p.basis.extension.pow(p.basis.omega, index+1)), nil
}
