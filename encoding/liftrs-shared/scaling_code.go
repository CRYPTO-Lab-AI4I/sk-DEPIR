package liftrs

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

// ScalingCodeHeader describes experimental mathematical code data for m=2.
// Header inspection validates representation only, never payload, provenance or
// code identities. Version1/2 are existing LRSprobe1/2; version3 stores explicit
// field/tower/generator identifiers and8-byte extension values at s17..20.
// Offsets/Bytes refer to the actual input representation, including its header.
type ScalingCodeHeader struct {
	Spec                                Spec
	Version, CountBytes, ExtensionBytes uint8
	K, N                                uint64
	Polynomial, Delta                   uint32
	Omega                               uint64
	POffset, SpectrumOffset, Bytes      uint64
}

// ScalingCodeLimits are mandatory admission limits before resident allocation
// or payload decoding. MaxPayloadBytes caps principal-array bytes (4(K+1)+wN),
// not total RSS: bounded64KiB I/O buffers, field metadata and runtime are extra.
// Verification uses the same logical extent caps but allocates no payload arrays.
type ScalingCodeLimits struct{ MaxElements, MaxPayloadBytes uint64 }

// ScalingCodeData holds immutable code-only coefficients with size-dependent
// slots. It is separate from ScalingOnlinePlan and does not admit a wider
// encoder. Base P coefficients remain uint32; spectra use uint32 through s16
// and uint64 at s17..20. Callers must obtain data from verified preprocessing;
// digests/range checks do not prove code identities or PIR security.
type ScalingCodeData struct {
	header        ScalingCodeHeader
	field         wideOrbitField
	p, spectrum32 []uint32
	spectrum64    []uint64
}

// Header returns a copy of the decoded representation metadata.
func (d *ScalingCodeData) Header() ScalingCodeHeader { return d.header }

// ResidentBytes counts principal arrays only, excluding metadata/runtime overhead.
func (d *ScalingCodeData) ResidentBytes() uint64 {
	return 4*uint64(len(d.p)+len(d.spectrum32)) + 8*uint64(len(d.spectrum64))
}

// SystematicPosition maps a zero-based message index using the stored canonical
// primitive generator. Both exponent and extension coordinate retain64 bits.
func (d *ScalingCodeData) SystematicPosition(index uint64) (uint64, error) {
	if index >= d.header.K {
		return 0, fmt.Errorf("systematic index outside message")
	}
	return d.field.pow(d.header.Omega, index+1), nil
}

func scalingCodeShape(s, t int, k, n uint64) bool {
	return s >= 2 && s <= 20 && t >= 1 && t < (1<<uint(s))-1 && n == uint64(1)<<uint(2*s) && k >= 1 && k <= n/2
}
func scalingCodeLayout(h *ScalingCodeHeader) {
	h.POffset = 25
	if h.Version == 2 {
		h.POffset = 33
	}
	if h.Version == 3 {
		h.POffset = 53
	}
	h.SpectrumOffset = h.POffset + 4*(h.K+1)
	h.Bytes = h.SpectrumOffset + uint64(h.ExtensionBytes)*h.N
}
func canonicalScalingCodeHeader(s, t int, k, n uint64) (ScalingCodeHeader, wideOrbitField, error) {
	var h ScalingCodeHeader
	if !scalingCodeShape(s, t, k, n) {
		return h, wideOrbitField{}, fmt.Errorf("invalid scaling code shape")
	}
	poly, err := BinaryPolynomial(s)
	if err != nil {
		return h, wideOrbitField{}, err
	}
	f, err := newWideOrbitField(s, poly)
	if err != nil {
		return h, f, err
	}
	omega, err := f.generator()
	if err != nil {
		return h, f, err
	}
	h = ScalingCodeHeader{Spec: Spec{s, 2, t}, Version: 1, CountBytes: 4, ExtensionBytes: 4, K: k, N: n, Polynomial: poly, Delta: f.tower.Delta(), Omega: omega}
	if s >= 16 {
		h.Version = 2
		h.CountBytes = 8
	}
	if s >= 17 {
		h.Version = 3
		h.ExtensionBytes = 8
	}
	scalingCodeLayout(&h)
	return h, f, nil
}

// ReadScalingCodeHeader reads at most53 bytes and allocates no domain arrays.
// It recognizes old files and checks new explicit identifiers against current
// canonical coordinates. This is unverified metadata, not a usable loaded plan.
func ReadScalingCodeHeader(source io.Reader) (ScalingCodeHeader, error) {
	h, _, err := readScalingCodeHeader(source)
	return h, err
}
func readScalingCodeHeader(source io.Reader) (ScalingCodeHeader, wideOrbitField, error) {
	var h ScalingCodeHeader
	var magic [9]byte
	if _, err := io.ReadFull(source, magic[:]); err != nil {
		return h, wideOrbitField{}, err
	}
	switch string(magic[:]) {
	case "LRSprobe1":
		h.Version = 1
		h.CountBytes = 4
	case "LRSprobe2":
		h.Version = 2
		h.CountBytes = 8
	case "LRSprobe3":
		h.Version = 3
		h.CountBytes = 8
	default:
		return h, wideOrbitField{}, fmt.Errorf("bad scaling code magic")
	}
	var spec [2]uint32
	if err := binary.Read(source, binary.LittleEndian, &spec); err != nil {
		return h, wideOrbitField{}, err
	}
	var counts [2]uint64
	if h.CountBytes == 4 {
		var narrow [2]uint32
		if err := binary.Read(source, binary.LittleEndian, &narrow); err != nil {
			return h, wideOrbitField{}, err
		}
		counts = [2]uint64{uint64(narrow[0]), uint64(narrow[1])}
	} else if err := binary.Read(source, binary.LittleEndian, &counts); err != nil {
		return h, wideOrbitField{}, err
	}
	s, t := int(spec[0]), int(spec[1])
	if !scalingCodeShape(s, t, counts[0], counts[1]) || (h.Version < 3 && s > 16) || (h.Version == 3 && s < 17) {
		return h, wideOrbitField{}, fmt.Errorf("invalid scaling code width/shape")
	}
	canonical, f, err := canonicalScalingCodeHeader(s, t, counts[0], counts[1])
	if err != nil {
		return h, f, err
	}
	if h.Version == 3 {
		var ids [2]uint32
		var omega uint64
		var widths [4]byte
		if err := binary.Read(source, binary.LittleEndian, &ids); err != nil {
			return h, f, err
		}
		if err := binary.Read(source, binary.LittleEndian, &omega); err != nil {
			return h, f, err
		}
		if _, err := io.ReadFull(source, widths[:]); err != nil {
			return h, f, err
		}
		if ids != [2]uint32{canonical.Polynomial, canonical.Delta} || omega != canonical.Omega || widths != [4]byte{8, 4, 8, 0} {
			return h, f, fmt.Errorf("noncanonical scaling code identifiers/slots")
		}
	}
	version, countBytes := h.Version, h.CountBytes
	h = canonical
	h.Version = version
	h.CountBytes = countBytes
	scalingCodeLayout(&h)
	return h, f, nil
}

func scalingCodeAdmission(h ScalingCodeHeader, limits ScalingCodeLimits, resident bool) error {
	bytes := h.Bytes - h.POffset
	if limits.MaxElements == 0 || limits.MaxPayloadBytes == 0 || h.N > limits.MaxElements || bytes > limits.MaxPayloadBytes {
		return fmt.Errorf("scaling code count/byte cap exceeded")
	}
	if resident && (h.N > uint64(^uint(0)>>1)/uint64(h.ExtensionBytes) || h.K+1 > uint64(^uint(0)>>1)/4) {
		return fmt.Errorf("scaling code arrays exceed host width")
	}
	return nil
}

// LoadScalingCode verifies a mandatory SHA-256 supplied by a trusted manifest,
// validates every element, and reads one resident copy using width-dependent
// slots. Input must end exactly after the spectrum. Caller budgets must include
// runtime/headroom beyond the principal-array cap. No existing cache is changed.
func LoadScalingCode(source io.Reader, digest string, limits ScalingCodeLimits) (*ScalingCodeData, error) {
	d, _, err := readScalingCode(source, digest, limits, true)
	return d, err
}

// VerifyScalingCode validates the same header/payload/digest with bounded I/O
// memory and no resident arrays. SHA identity is not mathematical code validity.
func VerifyScalingCode(source io.Reader, digest string, limits ScalingCodeLimits) (ScalingCodeHeader, error) {
	_, h, err := readScalingCode(source, digest, limits, false)
	return h, err
}
func readScalingCode(source io.Reader, digest string, limits ScalingCodeLimits, resident bool) (*ScalingCodeData, ScalingCodeHeader, error) {
	var empty ScalingCodeHeader
	expected, err := hex.DecodeString(digest)
	if err != nil || len(expected) != sha256.Size {
		return nil, empty, fmt.Errorf("trusted scaling code SHA-256 required")
	}
	hash := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(source, hash), 64<<10)
	h, f, err := readScalingCodeHeader(r)
	if err != nil {
		return nil, empty, err
	}
	if err = scalingCodeAdmission(h, limits, resident); err != nil {
		return nil, empty, err
	}
	d := &ScalingCodeData{header: h, field: f}
	if resident {
		d.p = make([]uint32, int(h.K+1))
	}
	first, err := readScalingCodeValues(r, h.K+1, 4, uint64(1)<<uint(h.Spec.FieldDegree), d.p, nil)
	if err != nil {
		return nil, empty, err
	}
	if first != 1 {
		return nil, empty, fmt.Errorf("invalid P constant")
	}
	if resident {
		if h.ExtensionBytes == 4 {
			d.spectrum32 = make([]uint32, int(h.N))
		} else {
			d.spectrum64 = make([]uint64, int(h.N))
		}
	}
	if _, err = readScalingCodeValues(r, h.N, h.ExtensionBytes, uint64(1)<<uint(2*h.Spec.FieldDegree), d.spectrum32, d.spectrum64); err != nil {
		return nil, empty, err
	}
	if _, err = r.ReadByte(); err != io.EOF {
		return nil, empty, fmt.Errorf("trailing scaling code data or read error")
	}
	if subtle.ConstantTimeCompare(hash.Sum(nil), expected) != 1 {
		return nil, empty, fmt.Errorf("scaling code digest differs from manifest")
	}
	if !resident {
		return nil, h, nil
	}
	return d, h, nil
}

// The bounded decoder also serves streaming verification. All counts/offsets
// stay uint64; a host index is formed only for admitted resident slices.
func readScalingCodeValues(r io.Reader, count uint64, width uint8, limit uint64, narrow []uint32, wide []uint64) (uint64, error) {
	if width != 4 && width != 8 {
		return 0, fmt.Errorf("invalid scaling slot width")
	}
	var block [65536]byte
	var first uint64
	for start := uint64(0); start < count; {
		chunk := uint64(len(block)) / uint64(width)
		if chunk > count-start {
			chunk = count - start
		}
		if _, err := io.ReadFull(r, block[:int(chunk)*int(width)]); err != nil {
			return 0, err
		}
		for j := 0; j < int(chunk); j++ {
			offset := j * int(width)
			x := uint64(0)
			if width == 4 {
				x = uint64(binary.LittleEndian.Uint32(block[offset:]))
			} else {
				x = binary.LittleEndian.Uint64(block[offset:])
			}
			if x >= limit {
				return 0, fmt.Errorf("scaling code element outside field")
			}
			if start == 0 && j == 0 {
				first = x
			}
			if narrow != nil {
				if x > uint64(^uint32(0)) {
					return 0, fmt.Errorf("narrow scaling code overflow")
				}
				narrow[int(start)+j] = uint32(x)
			}
			if wide != nil {
				wide[int(start)+j] = x
			}
		}
		start += chunk
	}
	return first, nil
}

// WriteScalingCode streams canonical size-dependent storage: old byte-compatible
// headers/32-bit values through s16, or LRSprobe3 with explicit identifiers and
// 64-bit spectrum at s17..20. It checks ranges, not code identities. The caller
// owns atomic persistence, the output digest and provenance. Errors may leave a
// partial destination; never overwrite a retained cache without atomic handling.
func WriteScalingCode(destination io.Writer, d *ScalingCodeData) error {
	if d == nil {
		return fmt.Errorf("nil scaling code")
	}
	h, _, err := canonicalScalingCodeHeader(d.header.Spec.FieldDegree, d.header.Spec.CurveDegree, d.header.K, d.header.N)
	if err != nil {
		return err
	}
	if uint64(len(d.p)) != h.K+1 || h.ExtensionBytes == 4 && (uint64(len(d.spectrum32)) != h.N || len(d.spectrum64) != 0) || h.ExtensionBytes == 8 && (uint64(len(d.spectrum64)) != h.N || len(d.spectrum32) != 0) {
		return fmt.Errorf("scaling code array shape")
	}
	if d.header.Spec != h.Spec || d.header.Polynomial != h.Polynomial || d.header.Delta != h.Delta || d.header.Omega != h.Omega {
		return fmt.Errorf("noncanonical scaling code metadata")
	}
	if d.p[0] != 1 {
		return fmt.Errorf("invalid P constant")
	}
	w := bufio.NewWriterSize(destination, 64<<10)
	if err = writeScalingCodeHeader(w, h); err != nil {
		return err
	}
	if err = writeScalingCodeValues(w, 4, uint64(1)<<uint(h.Spec.FieldDegree), d.p, nil); err != nil {
		return err
	}
	if err = writeScalingCodeValues(w, h.ExtensionBytes, uint64(1)<<uint(2*h.Spec.FieldDegree), d.spectrum32, d.spectrum64); err != nil {
		return err
	}
	return w.Flush()
}
func writeScalingCodeHeader(w io.Writer, h ScalingCodeHeader) error {
	if h.Version < 3 {
		return writeScalingOnlineHeader(w, h.Spec.FieldDegree, h.Spec.CurveDegree, h.K, h.N)
	}
	if _, err := io.WriteString(w, "LRSprobe3"); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, [2]uint32{uint32(h.Spec.FieldDegree), uint32(h.Spec.CurveDegree)}); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, [2]uint64{h.K, h.N}); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, [2]uint32{h.Polynomial, h.Delta}); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, h.Omega); err != nil {
		return err
	}
	_, err := w.Write([]byte{8, 4, 8, 0})
	return err
}
func writeScalingCodeValues(w io.Writer, width uint8, limit uint64, narrow []uint32, wide []uint64) error {
	if width != 4 && width != 8 {
		return fmt.Errorf("invalid scaling slot width")
	}
	count := len(narrow)
	if wide != nil {
		count = len(wide)
	}
	var block [65536]byte
	for start := 0; start < count; {
		chunk := len(block) / int(width)
		if chunk > count-start {
			chunk = count - start
		}
		for j := 0; j < chunk; j++ {
			x := uint64(0)
			if wide != nil {
				x = wide[start+j]
			} else {
				x = uint64(narrow[start+j])
			}
			if x >= limit {
				return fmt.Errorf("scaling code element outside field")
			}
			if width == 4 {
				if x > uint64(^uint32(0)) {
					return fmt.Errorf("narrow scaling code overflow")
				}
				binary.LittleEndian.PutUint32(block[4*j:], uint32(x))
			} else {
				binary.LittleEndian.PutUint64(block[8*j:], x)
			}
		}
		if n, err := w.Write(block[:chunk*int(width)]); err != nil {
			return err
		} else if n != chunk*int(width) {
			return io.ErrShortWrite
		}
		start += chunk
	}
	return nil
}
