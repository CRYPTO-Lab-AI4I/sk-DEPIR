package liftrs

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// LoadScalingOnline reads experimental LRSprobe1/2 minimal plans with bounded
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
	if string(magic[:]) != "LRSprobe1" && string(magic[:]) != "LRSprobe2" {
		return nil, fmt.Errorf("bad probe magic")
	}
	var spec [2]uint32
	if err := binary.Read(r, binary.LittleEndian, &spec); err != nil {
		return nil, err
	}
	var counts [2]uint64
	if string(magic[:]) == "LRSprobe1" {
		var narrow [2]uint32
		if err := binary.Read(r, binary.LittleEndian, &narrow); err != nil {
			return nil, err
		}
		counts = [2]uint64{uint64(narrow[0]), uint64(narrow[1])}
	} else if err := binary.Read(r, binary.LittleEndian, &counts); err != nil {
		return nil, err
	}
	s, t := int(spec[0]), int(spec[1])
	// Check wide counts before converting to host lengths or allocating arrays.
	if !validScalingOnlineCounts(s, t, counts[0], counts[1], maxElements) {
		return nil, fmt.Errorf("invalid online probe shape/limit")
	}
	k, n := int(counts[0]), int(counts[1])
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
				if uint64(x) >= limit {
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

// WriteScalingOnline streams a verified plan without changing its field values.
// s<=15 uses byte-compatible LRSprobe1 with 32-bit counts; s=16 uses LRSprobe2
// with 64-bit K/N. The caller owns atomic persistence and digest/provenance storage.
// As with loading, this experimental format is not a public distribution contract.
func WriteScalingOnline(destination io.Writer, p *ScalingOnlinePlan) error {
	if p == nil || !validScalingOnlineShape(p.basis.spec.FieldDegree, p.basis.spec.CurveDegree, p.k, p.Length(), p.Length()) ||
		len(p.p) != p.k+1 || len(p.spectrum) != p.Length() {
		return fmt.Errorf("invalid online plan")
	}
	w := bufio.NewWriterSize(destination, 64<<10)
	if err := writeScalingOnlineHeader(w, p.FieldDegree(), p.basis.spec.CurveDegree, uint64(p.k), uint64(p.Length())); err != nil {
		return err
	}
	var block [65536]byte
	for _, values := range [][]uint32{p.p, p.spectrum} {
		for start := 0; start < len(values); start += len(block) / 4 {
			count := minInt(len(block)/4, len(values)-start)
			for j := 0; j < count; j++ {
				binary.LittleEndian.PutUint32(block[4*j:], values[start+j])
			}
			if _, err := w.Write(block[:4*count]); err != nil {
				return err
			}
		}
	}
	return w.Flush()
}

func writeScalingOnlineHeader(w io.Writer, s, t int, k, n uint64) error {
	magic := "LRSprobe1"
	if s >= 16 {
		magic = "LRSprobe2"
	}
	if _, err := io.WriteString(w, magic); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, [2]uint32{uint32(s), uint32(t)}); err != nil {
		return err
	}
	if s <= 15 {
		return binary.Write(w, binary.LittleEndian, [2]uint32{uint32(k), uint32(n)})
	}
	return binary.Write(w, binary.LittleEndian, [2]uint64{k, n})
}

// Validate counts before any allocation. The four-byte value representation
// supports s<=16; the count and the host's byte offsets must not overflow.
func validScalingOnlineShape(s, t, k, n, maxElements int) bool {
	return k >= 1 && n >= 1 && validScalingOnlineCounts(s, t, uint64(k), uint64(n), maxElements)
}

func validScalingOnlineCounts(s, t int, k, n uint64, maxElements int) bool {
	return s >= 2 && s <= 16 && t >= 1 && t < (1<<uint(s))-1 && maxElements > 0 &&
		n == uint64(1)<<uint(2*s) && n <= uint64(maxElements) && n <= uint64(^uint(0)>>1)/4 && k >= 1 && k <= n/2
}
