package liftrs

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// SupportRecord is a trusted manifest for a separate raw little-endian payload.
// Coordinates use4-byte slots through s16 and8 thereafter. JSON counts are64-bit.
// A self-computed digest is not evidence of exact dimension/support correctness;
// obtain this record from certified discovery or an independently trusted source.
type SupportRecord struct {
	Version      int               `json:"version"`
	Parameters   SupportParameters `json:"parameters"`
	ElementBytes uint8             `json:"element_bytes"`
	Bytes        uint64            `json:"bytes"`
	SHA256       string            `json:"sha256"`
}

func supportWidth(s int) uint8 {
	if s <= 16 {
		return 4
	}
	return 8
}

// Write streams public support without copying the full array. The caller owns
// atomic publication and must persist the returned manifest/provenance. A write
// failure can leave a partial destination; no completed record is returned.
func (s *CodeSupport) Write(destination io.Writer) (SupportRecord, error) {
	if s == nil || destination == nil || s.parameters.K == 0 || s.parameters.K != uint64(len(s.exponents)) || s.parameters.Spec.FieldDegree < 2 || s.parameters.Spec.FieldDegree > 17 {
		return SupportRecord{}, fmt.Errorf("support writer required")
	}
	h := s.parameters
	width := supportWidth(h.Spec.FieldDegree)
	hash := sha256.New()
	writer := io.MultiWriter(destination, hash)
	var block [65536]byte
	reporter := newPreparationReporter(nil)
	for first := 0; first < len(s.exponents); {
		take := minInt(len(block)/int(width), len(s.exponents)-first)
		for j := 0; j < take; j++ {
			v := uint64(s.exponents[first+j])
			offset := j * int(width)
			if width == 4 {
				if v > uint64(^uint32(0)) {
					return SupportRecord{}, fmt.Errorf("support coordinate narrowing")
				}
				binary.LittleEndian.PutUint32(block[offset:], uint32(v))
			} else {
				binary.LittleEndian.PutUint64(block[offset:], v)
			}
		}
		if _, err := writer.Write(block[:take*int(width)]); err != nil {
			return SupportRecord{}, err
		}
		first += take
		reporter.notice("write-support", uint64(first), h.K, "exponents")
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	s.setChecksum(digest)
	return SupportRecord{Version: 1, Parameters: h, ElementBytes: width, Bytes: h.K * uint64(width), SHA256: digest}, nil
}
func validSupportParameters(info SupportParameters, limits PreprocessingLimits) error {
	s := info.Spec.FieldDegree
	if s < 2 || s > 17 || info.Spec.Dimension != 2 || info.K < 1 || info.K >= info.N || info.Samples < 0 || limits.MaxElements < info.N || limits.MaxHostArrayBytes == 0 || info.N > uint64(^uint(0)>>1) || info.K > uint64(^uint(0)>>1)/8 || info.K*8 > limits.MaxHostArrayBytes || info.N/8 > limits.MaxHostArrayBytes {
		return fmt.Errorf("support shape/host limits")
	}
	h, _, err := canonicalScalingCodeHeader(s, info.Spec.CurveDegree, 1, info.N)
	if err != nil {
		return err
	}
	if info.Polynomial != h.Polynomial || info.Delta != h.Delta || info.Omega != h.Omega {
		return fmt.Errorf("support field representation mismatch")
	}
	return nil
}
func expectedDigest(value string) ([]byte, error) {
	expected, err := hex.DecodeString(value)
	if err != nil || len(expected) != sha256.Size {
		return nil, fmt.Errorf("trusted SHA-256 required")
	}
	return expected, nil
}

// ReadCodeSupport checks a trusted manifest, caps, every coordinate, exact EOF,
// digest, ordering and Frobenius closure. It does not redo expensive certified
// discovery; the manifest must come from that process. Principal-array limits
// exclude their combined live sizes, allocator/runtime and filesystem cache.
func ReadCodeSupport(source io.Reader, record SupportRecord, limits PreprocessingLimits) (*CodeSupport, error) {
	if err := validSupportParameters(record.Parameters, limits); err != nil {
		return nil, err
	}
	expected, err := expectedDigest(record.SHA256)
	if err != nil {
		return nil, err
	}
	info := record.Parameters
	width := supportWidth(info.Spec.FieldDegree)
	if record.Version != 1 || record.ElementBytes != width || record.Bytes != info.K*uint64(width) {
		return nil, fmt.Errorf("support manifest width/size/version")
	}
	hash := sha256.New()
	reader := bufio.NewReaderSize(io.TeeReader(source, hash), 65536)
	indices := make([]int, int(info.K))
	var block [65536]byte
	reporter := newPreparationReporter(nil)
	for first := 0; first < len(indices); {
		take := minInt(len(block)/int(width), len(indices)-first)
		if _, err = io.ReadFull(reader, block[:take*int(width)]); err != nil {
			return nil, err
		}
		for j := 0; j < take; j++ {
			offset := j * int(width)
			var value uint64
			if width == 4 {
				value = uint64(binary.LittleEndian.Uint32(block[offset:]))
			} else {
				value = binary.LittleEndian.Uint64(block[offset:])
			}
			if value >= info.N-1 || value > uint64(^uint(0)>>1) {
				return nil, fmt.Errorf("support coordinate outside domain")
			}
			indices[first+j] = int(value)
		}
		first += take
		reporter.notice("read-support", uint64(first), info.K, "exponents")
	}
	if _, err = reader.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("trailing support bytes/read error")
	}
	if subtle.ConstantTimeCompare(hash.Sum(nil), expected) != 1 {
		return nil, fmt.Errorf("support checksum mismatch")
	}
	if err = validateExponentSupport(indices, int(info.N), uint64(1)<<uint(info.Spec.FieldDegree)); err != nil {
		return nil, err
	}
	return &CodeSupport{parameters: info, exponents: indices, checksum: hex.EncodeToString(hash.Sum(nil))}, nil
}

// RootRecord binds normalized root coefficients to a persisted support identity.
// This is a trusted restart artifact, not a mathematical proof or codeword file.
type RootRecord struct {
	Version       int               `json:"version"`
	Parameters    SupportParameters `json:"parameters"`
	SupportSHA256 string            `json:"support_sha256"`
	Bytes         uint64            `json:"bytes"`
	SHA256        string            `json:"sha256"`
}

// RootCheckpoint is immutable normalized P, bound to a trusted source support.
// It is created by preparation or ReadPreparationRoot; callers cannot inject
// unbound coefficient arrays. The full P*G identity is still rechecked on reuse.
type RootCheckpoint struct {
	parameters SupportParameters
	supportSHA string
	values     []uint32
}

func sameSupportCode(a, b SupportParameters) bool { a.Samples = 0; b.Samples = 0; return a == b }

func (r *RootCheckpoint) matches(support *CodeSupport) error {
	if r == nil || support == nil || !sameSupportCode(r.parameters, support.parameters) || r.supportSHA == "" || r.supportSHA != support.Checksum() || uint64(len(r.values)) != r.parameters.K+1 {
		return fmt.Errorf("root checkpoint differs from trusted support")
	}
	return nil
}
func (r *RootCheckpoint) Write(destination io.Writer) (RootRecord, error) {
	if r == nil || destination == nil || r.supportSHA == "" {
		return RootRecord{}, fmt.Errorf("bound root checkpoint/writer required")
	}
	hash := sha256.New()
	if err := writeScalingCodeValues(io.MultiWriter(destination, hash), 4, uint64(1)<<uint(r.parameters.Spec.FieldDegree), r.values, nil); err != nil {
		return RootRecord{}, err
	}
	return RootRecord{Version: 1, Parameters: r.parameters, SupportSHA256: r.supportSHA, Bytes: 4 * uint64(len(r.values)), SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}

// ReadPreparationRoot verifies source binding, size, digest, base-field values,
// normalization and EOF without changing retained files. A matching trusted
// manifest is mandatory; mathematical generator verification happens in Prepare.
func ReadPreparationRoot(source io.Reader, support *CodeSupport, record RootRecord, limits PreprocessingLimits) (*RootCheckpoint, error) {
	if support == nil {
		return nil, fmt.Errorf("trusted support required")
	}
	if err := validSupportParameters(record.Parameters, limits); err != nil {
		return nil, err
	}
	if record.Version != 1 || !sameSupportCode(record.Parameters, support.parameters) || strings.ToLower(record.SupportSHA256) != support.Checksum() || support.Checksum() == "" || record.Bytes != 4*(record.Parameters.K+1) || record.Bytes > limits.MaxHostArrayBytes {
		return nil, fmt.Errorf("root manifest shape/source/cap")
	}
	expected, err := expectedDigest(record.SHA256)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	reader := bufio.NewReaderSize(io.TeeReader(source, hash), 65536)
	values := make([]uint32, int(record.Parameters.K+1))
	if _, err = readScalingCodeValues(reader, record.Parameters.K+1, 4, uint64(1)<<uint(record.Parameters.Spec.FieldDegree), values, nil); err != nil {
		return nil, err
	}
	if values[0] != 1 {
		return nil, fmt.Errorf("root normalization")
	}
	if _, err = reader.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("trailing root bytes/read error")
	}
	if subtle.ConstantTimeCompare(hash.Sum(nil), expected) != 1 {
		return nil, fmt.Errorf("root checksum mismatch")
	}
	return &RootCheckpoint{parameters: record.Parameters, supportSHA: strings.ToLower(record.SupportSHA256), values: values}, nil
}
