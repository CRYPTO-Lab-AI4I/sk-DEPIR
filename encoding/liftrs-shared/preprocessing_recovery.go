package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/bits"
)

// PreprocessingRecovery enables optional, synchronous safe-boundary persistence.
// Save receives borrowed, read-only arrays valid only until it returns: stream
// them with Write, never retain the checkpoint itself. Returning an error stops
// work at that boundary. Loaded checkpoints must come from a trusted local cache;
// digests detect damage, not a malicious substitute for certified discovery.
// With nil Recovery there is no checkpoint I/O or extra transform work.
type PreprocessingRecovery struct {
	Discovery, Preparation *PreprocessingCheckpoint
	Save                   func(*PreprocessingCheckpoint) error
}

// ProgressRecord describes a partial computation, never usable encoding data.
// Kind is discovery (union bitmap), roots (product frontier), or reciprocal
// (completed Newton prefix). Preparation records bind to the persisted support.
// Sections gives coefficient counts, preserving empty binary-carry levels.
type ProgressRecord struct {
	Version       int               `json:"version"`
	Kind          string            `json:"kind"`
	Parameters    SupportParameters `json:"parameters"`
	SupportSHA256 string            `json:"support_sha256,omitempty"`
	Next          uint64            `json:"next"`
	Count         uint64            `json:"count"`
	Sections      []uint64          `json:"sections,omitempty"`
	Bytes         uint64            `json:"bytes"`
	SHA256        string            `json:"sha256"`
}

// PreprocessingCheckpoint holds a recovery state. Fields are private to prevent
// construction from unchecked arrays. Write uses bounded buffers; it does not
// copy a grid, force a product-tree merge, or save live FFT workspace.
type PreprocessingCheckpoint struct {
	record   ProgressRecord
	bitmap   exponentSet
	sections [][]uint32
}

// Record returns metadata by value, including an independent section-length list.
func (c *PreprocessingCheckpoint) Record() ProgressRecord {
	if c == nil {
		return ProgressRecord{}
	}
	r := c.record
	r.Version = 1
	r.Bytes = uint64(len(c.bitmap)) * 8
	r.Sections = nil
	for _, p := range c.sections {
		r.Sections = append(r.Sections, uint64(len(p)))
		r.Bytes += uint64(len(p)) * 4
	}
	return r
}
func (c *PreprocessingCheckpoint) Write(w io.Writer) (ProgressRecord, error) {
	if c == nil || w == nil {
		return ProgressRecord{}, fmt.Errorf("checkpoint/writer required")
	}
	r := c.Record()
	h := sha256.New()
	w = io.MultiWriter(w, h)
	var block [65536]byte
	for first := 0; first < len(c.bitmap); {
		take := minInt(len(block)/8, len(c.bitmap)-first)
		for j, v := range c.bitmap[first : first+take] {
			binary.LittleEndian.PutUint64(block[8*j:], v)
		}
		if _, err := w.Write(block[:take*8]); err != nil {
			return ProgressRecord{}, err
		}
		first += take
	}
	for _, p := range c.sections {
		if err := writeScalingCodeValues(w, 4, uint64(1)<<uint(r.Parameters.Spec.FieldDegree), p, nil); err != nil {
			return ProgressRecord{}, err
		}
	}
	r.SHA256 = hex.EncodeToString(h.Sum(nil))
	return r, nil
}

// ReadPreprocessingCheckpoint authenticates bytes against a trusted manifest,
// validates shape, field values and allocation bounds, and rejects trailing data.
// Resume still validates source identity and the final support/P*G identity.
func ReadPreprocessingCheckpoint(src io.Reader, r ProgressRecord, limits PreprocessingLimits) (*PreprocessingCheckpoint, error) {
	if err := validSupportParameters(r.Parameters, limits); err != nil {
		return nil, err
	}
	if _, err := expectedDigest(r.SHA256); err != nil {
		return nil, err
	}
	if r.Version != 1 || len(r.Sections) > 66 {
		return nil, fmt.Errorf("checkpoint version/sections")
	}
	info := r.Parameters
	var words uint64
	for _, n := range r.Sections {
		if n > info.N || n*4 > limits.MaxHostArrayBytes || words > info.N-n {
			return nil, fmt.Errorf("checkpoint array cap")
		}
		words += n
	}
	bitmapBytes := uint64(0)
	switch r.Kind {
	case "discovery":
		bitmapBytes = (info.N + 63) / 64 * 8
		if len(r.Sections) != 0 || info.Samples < 1 || r.Count > info.K || r.Next != 0 || r.SupportSHA256 != "" {
			return nil, fmt.Errorf("discovery checkpoint shape")
		}
	case "roots":
		if len(r.Sections) < 1 || r.Sections[0]%3 != 0 || r.Sections[0]/3 >= uint64(minInt(32768, int(info.K))) || r.Next == 0 || r.Next > info.K || r.Next%65536 != 0 && r.Next != info.K {
			return nil, fmt.Errorf("root frontier shape")
		}
	case "reciprocal":
		target := info.N - info.K
		if len(r.Sections) != 1 || r.Next == 0 || r.Next > target || r.Sections[0] != r.Next || r.Next != target && r.Next&(r.Next-1) != 0 {
			return nil, fmt.Errorf("reciprocal checkpoint shape")
		}
	default:
		return nil, fmt.Errorf("unknown checkpoint kind")
	}
	if r.Kind != "discovery" {
		if _, err := expectedDigest(r.SupportSHA256); err != nil {
			return nil, err
		}
	}
	if r.Bytes != bitmapBytes+4*words || bitmapBytes > limits.MaxHostArrayBytes {
		return nil, fmt.Errorf("checkpoint size/cap")
	}
	h := sha256.New()
	reader := io.TeeReader(src, h)
	c := &PreprocessingCheckpoint{record: r}
	var block [65536]byte
	if bitmapBytes != 0 {
		c.bitmap = make(exponentSet, bitmapBytes/8)
		var count uint64
		for first := 0; first < len(c.bitmap); {
			take := minInt(len(block)/8, len(c.bitmap)-first)
			if _, err := io.ReadFull(reader, block[:take*8]); err != nil {
				return nil, err
			}
			for j := 0; j < take; j++ {
				v := binary.LittleEndian.Uint64(block[8*j:])
				c.bitmap[first+j] = v
				count += uint64(bits.OnesCount64(v))
			}
			first += take
		}
		if count != r.Count || c.bitmap.has(int(info.N-1)) || info.N%64 != 0 && c.bitmap[len(c.bitmap)-1]>>uint(info.N%64) != 0 {
			return nil, fmt.Errorf("checkpoint support bitmap/count")
		}
	}
	q := uint64(1) << uint(info.Spec.FieldDegree)
	for _, n := range r.Sections {
		p := make([]uint32, int(n))
		c.sections = append(c.sections, p)
		for first := 0; first < len(p); {
			take := minInt(len(block)/4, len(p)-first)
			if _, err := io.ReadFull(reader, block[:take*4]); err != nil {
				return nil, err
			}
			for j := 0; j < take; j++ {
				v := binary.LittleEndian.Uint32(block[4*j:])
				if uint64(v) >= q {
					return nil, fmt.Errorf("checkpoint coefficient field")
				}
				p[first+j] = v
			}
			first += take
		}
	}
	var extra [1]byte
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		return nil, fmt.Errorf("checkpoint trailing bytes/read error")
	}
	if hex.EncodeToString(h.Sum(nil)) != r.SHA256 {
		return nil, fmt.Errorf("checkpoint checksum mismatch")
	}
	return c, nil
}

func (c *PreprocessingCheckpoint) matches(info SupportParameters, sha string, kind string) error {
	if c == nil || c.record.Kind != kind || !sameSupportCode(c.record.Parameters, info) || c.record.SupportSHA256 != sha {
		return fmt.Errorf("checkpoint code/source/kind mismatch")
	}
	return nil
}

// Restore without finishing/merging the tree. Arrays read from a checkpoint are
// immutable; only the small pending leaf buffer and outer level list are copied.
func (c *PreprocessingCheckpoint) restoreBuilder(f orbitField, support []int, n int, q uint64, notice preparationNotice) (*factorProductBuilder, error) {
	b := newFactorProductBuilder(f, minInt(32768, len(support)))
	b.levels = append([][]uint32(nil), c.sections[1:]...)
	for i, p := range b.levels {
		if len(p) == 0 {
			b.levels[i] = nil
		}
	}
	var degree uint64
	for _, p := range b.levels {
		if len(p) == 0 {
			continue
		}
		if len(p) < 2 || p[0] == 0 || p[len(p)-1] != 1 {
			return nil, fmt.Errorf("checkpoint product monicity")
		}
		degree += uint64(len(p) - 1)
	}
	pending := c.sections[0]
	for j := 0; j < len(pending); j += 3 {
		v := pending[j : j+3]
		d := int(v[2]) + 1
		if v[0] == 0 || d > 2 || d == 1 && v[1] != 1 {
			return nil, fmt.Errorf("checkpoint pending factor")
		}
		b.add(v[0], v[1], v[2], d)
		degree += uint64(d)
	}
	var expected uint64
	for i, index := range support[:int(c.record.Next)] {
		if notice != nil && i%65536 == 0 {
			notice("recovery-root-degree", uint64(i), c.record.Next, "exponents")
		}
		partner := int(uint64(index) * q % uint64(n-1))
		if partner < index {
			continue
		}
		expected++
		if partner != index {
			expected++
		}
	}
	if degree != expected {
		return nil, fmt.Errorf("checkpoint processed degree mismatch")
	}
	return b, nil
}
