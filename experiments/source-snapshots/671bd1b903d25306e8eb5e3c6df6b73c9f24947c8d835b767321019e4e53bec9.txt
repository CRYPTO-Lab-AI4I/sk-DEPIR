//go:build liftrs_cpp && cgo

package liftrs

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/bits"
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
	hash := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(file, hash), 1<<20)
	var magic [9]byte
	if _, err = io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if string(magic[:]) != "LRSprobe1" && string(magic[:]) != "LRSprobe2" {
		return nil, fmt.Errorf("bad probe magic")
	}
	var spec [2]uint32
	if err = binary.Read(r, binary.LittleEndian, &spec); err != nil {
		return nil, err
	}
	var counts [2]uint64
	header := uint64(33)
	if string(magic[:]) == "LRSprobe1" {
		var v [2]uint32
		if err = binary.Read(r, binary.LittleEndian, &v); err != nil {
			return nil, err
		}
		counts = [2]uint64{uint64(v[0]), uint64(v[1])}
		header = 25
	} else if err = binary.Read(r, binary.LittleEndian, &counts); err != nil {
		return nil, err
	}
	s, t := int(spec[0]), int(spec[1])
	if !validScalingOnlineCounts(s, t, counts[0], counts[1], maxElements) {
		return nil, fmt.Errorf("invalid online file shape/limit")
	}
	k, n := int(counts[0]), int(counts[1])
	l := 4
	for l < 2*k-1 {
		l *= 2
	}
	if l > n/2 {
		return nil, fmt.Errorf("file prefix operands exceed arena")
	}
	if uint64(info.Size()) != header+4*(counts[0]+1+counts[1]) {
		return nil, fmt.Errorf("code file length")
	}
	buffer := make([]byte, 1<<20)
	for section, count := range []int{k + 1, n} {
		limit := uint64(1) << uint(s)
		if section == 1 {
			limit = uint64(1) << uint(2*s)
		}
		for first := 0; first < count; {
			take := minInt(len(buffer)/4, count-first)
			if _, err = io.ReadFull(r, buffer[:4*take]); err != nil {
				return nil, err
			}
			if section == 0 && first == 0 && binary.LittleEndian.Uint32(buffer) != 1 {
				return nil, fmt.Errorf("invalid P constant")
			}
			// At s16 every uint32 is already a legal extension value.
			if limit < 1<<32 {
				for i := 0; i < take; i++ {
					if uint64(binary.LittleEndian.Uint32(buffer[4*i:])) >= limit {
						return nil, fmt.Errorf("file code element outside field")
					}
				}
			}
			first += take
		}
	}
	if _, err = r.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("trailing code data or read error")
	}
	if hex.EncodeToString(hash.Sum(nil)) != digest {
		return nil, fmt.Errorf("online file checksum differs from manifest")
	}
	after, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !sameCodeFile(info, after) {
		return nil, fmt.Errorf("code changed during verification")
	}
	polynomial, _ := BinaryPolynomial(s)
	base, extension, delta, omega := orbitFields(s, polynomial)
	plan := &ScalingOnlinePlan{basis: OrbitBasis{spec: Spec{s, 2, t}, polynomial: polynomial, delta: delta, omega: omega, base: base, extension: extension, extensionFFT: newSparseTransform(extension)}, k: k}
	cpu := &ScalingCPPEncoder{source: plan, workers: workers, plans: make(map[int]*matrixCPPPlan)}
	for _, count := range []int{l, n} {
		if cpu.plans[count] != nil {
			continue
		}
		v, err := newMatrixCPP(extension, bits.Len(uint(count))-1)
		if err != nil {
			cpu.Close()
			return nil, err
		}
		cpu.plans[count] = v
	}
	ok = true
	return &ScalingCPPFileEncoder{file: file, identity: after, cpu: cpu, prefix: l, pOffset: header, spectrumOffset: header + 4*uint64(k+1)}, nil
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
