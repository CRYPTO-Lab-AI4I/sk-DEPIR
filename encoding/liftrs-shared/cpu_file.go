//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
)

// ScalingCPPFileEncoder explicitly streams verified code from an immutable file.
// Calls serialize. The caller must prevent writes/truncation for its lifetime;
// descriptor/metadata checks detect common changes, not hostile concurrent writes.
// File cache is additional memory: lower private RSS is not a total RAM bound.
type ScalingCPPFileEncoder struct {
	mu                      sync.Mutex
	file                    *os.File
	identity                os.FileInfo
	cpu                     *ScalingCPPEncoder
	prefix                  int
	pOffset, spectrumOffset uint64
	arena, scratch          []uint32
	closed                  bool
}

// OpenScalingCPPFile checks the entire trusted SHA256 and field ranges before
// preparing bounded native metadata. The file must derive from verified code;
// a checksum authenticates byte identity, not mathematical validity. No P/spectrum
// arrays are retained. Existing LRSprobe1/2 bytes and resident APIs are unchanged.
// This fast explicit schedule requires 2L<=N and s<=16. First Encode allocates
// two N-element arrays; subsequent calls replace detached output scratch.
func OpenScalingCPPFile(path, digest string, maxElements, workers int) (*ScalingCPPFileEncoder, error) {
	if workers < 1 || workers > 64 {
		return nil, fmt.Errorf("invalid file encoder workers")
	}
	expected, err := hex.DecodeString(digest)
	if err != nil || len(expected) != sha256.Size {
		return nil, fmt.Errorf("expected SHA256 required")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			file.Close()
		}
	}()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("code backing must be a regular immutable file")
	}
	// Inspect the small header before streaming the payload: the existing file
	// encoder remains32-bit even though the common verifier also knows wide files.
	h, err := ReadScalingCodeHeader(file)
	if err != nil {
		return nil, err
	}
	s, t := h.Spec.FieldDegree, h.Spec.CurveDegree
	if h.ExtensionBytes != 4 || !validScalingOnlineCounts(s, t, h.K, h.N, maxElements) {
		return nil, fmt.Errorf("invalid online file shape/limit")
	}
	k, n := int(h.K), int(h.N)
	l := 4
	for l < 2*k-1 {
		l *= 2
	}
	if l > n/2 {
		return nil, fmt.Errorf("file prefix operands exceed arena")
	}
	if uint64(info.Size()) != h.Bytes {
		return nil, fmt.Errorf("code file length")
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if _, err = VerifyScalingCode(file, digest, ScalingCodeLimits{
		MaxElements: uint64(maxElements), MaxPayloadBytes: h.Bytes - h.POffset,
	}); err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !sameCodeFile(info, after) {
		return nil, fmt.Errorf("code changed during verification")
	}
	plan := &ScalingOnlinePlan{basis: scalingOnlineBasis(h), k: k}
	cpu, err := newScalingCPPEncoder(plan, workers, l)
	if err != nil {
		return nil, err
	}
	ok = true
	return &ScalingCPPFileEncoder{file: file, identity: after, cpu: cpu, prefix: l, pOffset: h.POffset, spectrumOffset: h.SpectrumOffset}, nil
}

func sameCodeFile(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

// Dimension is the verified input cardinality.
func (e *ScalingCPPFileEncoder) Dimension() int { return e.cpu.source.k }

// Length is the canonical output cardinality.
func (e *ScalingCPPFileEncoder) Length() int { return e.cpu.source.Length() }

// WorkspaceBytes excludes caller-owned detached words and the kernel file cache.
func (e *ScalingCPPFileEncoder) WorkspaceBytes() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return 4 * uint64(len(e.arena)+len(e.scratch))
}

// ScalingCPPFileStats reports native stages and logical file reads. ReadNS
// includes range checking; prefix read time also appears in BufferNS. ReadBytes
// is not physical disk traffic: the OS may serve reads from its page cache.
type ScalingCPPFileStats struct {
	ScalingCPPStats
	ReadNS     uint64 `json:"read_ns"`
	ReadBytes  uint64 `json:"read_bytes"`
	ReadChunks uint64 `json:"read_chunks"`
}

// EncodeWithStats selects file streaming and returns a detached owning uint32
// word. No descriptor is retained by native execution; Close waits for this call.
// Source metadata is checked before/after, under the immutable-file contract.
func (e *ScalingCPPFileEncoder) EncodeWithStats(message []uint32) ([]uint32, ScalingCPPFileStats, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.encodeFileLocked(message, false)
}

// Caller holds mu through use of borrowed packed scratch.
func (e *ScalingCPPFileEncoder) encodeFileLocked(message []uint32, packed bool) ([]uint32, ScalingCPPFileStats, error) {
	var stats ScalingCPPFileStats
	if e.closed {
		return nil, stats, ErrCPPClosed
	}
	p := e.cpu.source
	n := p.Length()
	if len(message) != p.k {
		return nil, stats, fmt.Errorf("file encoding message length")
	}
	before, err := e.file.Stat()
	if err != nil {
		return nil, stats, err
	}
	if !sameCodeFile(e.identity, before) {
		return nil, stats, fmt.Errorf("code backing changed")
	}
	if e.arena == nil {
		e.arena = make([]uint32, n)
	}
	if e.scratch == nil {
		e.scratch = make([]uint32, n)
	}
	pre, full := e.cpu.plans[e.prefix], e.cpu.plans[n]
	encode := cpp.EncodeOnlineFile
	if packed {
		encode = cpp.EncodeOnlineFilePacked
	}
	native, err := encode(pre.rows, pre.columns, full.rows, full.columns, pre.split, full.split, e.cpu.workers, p.basis.omega, message, e.arena, e.scratch, e.prefix, cpp.OnlineFile{Descriptor: int(e.file.Fd()), POffset: e.pOffset, SpectrumOffset: e.spectrumOffset, TileElements: minInt(n, 1<<26)})
	if err != nil {
		return nil, stats, err
	}
	after, err := e.file.Stat()
	if err != nil {
		return nil, stats, err
	}
	if !sameCodeFile(e.identity, after) {
		return nil, stats, fmt.Errorf("code backing changed during encoding")
	}
	stats.ScalingCPPStats = ScalingCPPStats{Calls: 5, BoundaryCalls: 1, TaylorNS: native.TaylorNS, ColumnsNS: native.ColumnsNS, RowsNS: native.RowsNS, BufferNS: native.BufferNS, PointwiseNS: native.PointwiseNS, AssemblyNS: native.AssemblyNS}
	stats.ReadNS = native.ReadNS
	stats.ReadBytes = native.ReadBytes
	stats.ReadChunks = native.ReadChunks
	word := e.scratch
	if !packed {
		e.scratch = nil
	}
	return word, stats, nil
}

// Close releases descriptor, plans and privately owned arrays; returned words
// remain valid. The immutable source file is never deleted or modified.
func (e *ScalingCPPFileEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	e.cpu.Close()
	e.arena = nil
	e.scratch = nil
	return e.file.Close()
}
