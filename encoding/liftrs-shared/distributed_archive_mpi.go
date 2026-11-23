//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// Read/hash the immutable archive once on rank0 and route bounded owner buckets.
// All source and payload checks finish before any root/GPU allocation. Checksum
// provenance is caller-trusted;this detects corruption,not malicious rewriting.
func loadRankDiscovery(w *mpi.World, o rankWorkflowOptions, r rankWorkflowRecovery, emit func(map[string]any)) (result discoveryResult, err error) {
	if w == nil || emit == nil || r.Known != 0 || r.Label != fmt.Sprintf("s%d", o.Spec.FieldDegree) {
		return result, fmt.Errorf("cached discovery admission")
	}
	source, err := readWorkflowCheckpointSource(r.Directory, r.Label, r.SourceSHA)
	if err != nil {
		return result, err
	}
	h := source.Header
	if h.Spec != o.Spec || source.Ranks != w.Size || r.Known != 0 {
		return result, fmt.Errorf("cached discovery source binding")
	}
	if o.Discovery.AdmitCode != nil {
		if err = o.Discovery.AdmitCode(h); err != nil {
			return result, err
		}
	}
	if o.Discovery.AdmitHost != nil {
		p := hostInventory("cached-discovery", map[string]uint64{"routed_capacity_allowance": 16 * o.Discovery.MaxRootWords, "archive_tiles": 3 * rankArchiveTile * 8, "codec_and_counts": 65536})
		if err = o.Discovery.AdmitHost(p); err != nil {
			return result, err
		}
	}
	var file *os.File
	hasher := sha256.New()
	bytes := make([]byte, 8*rankArchiveTile)
	raw := make([]uint64, rankArchiveTile)
	routed := make([]uint64, rankArchiveTile)
	if w.Rank == 0 {
		file, err = os.Open(filepath.Join(r.Directory, r.Label+"-support.bin"))
		if err != nil {
			return result, err
		}
		defer file.Close()
	}
	roots := []int{}
	last := time.Time{}
	previous := uint64(0)
	for offset := uint64(0); offset < h.K; {
		take := min(uint64(rankArchiveTile), h.K-offset)
		counts := make([]uint64, w.Size)
		starts := make([]uint64, w.Size+1)
		if w.Rank == 0 {
			data := bytes[:8*take]
			if err = readArchiveValues(file, h.N, &previous, offset > 0, data, raw[:take]); err != nil {
				return result, err
			}
			hasher.Write(data)
			for _, v := range raw[:take] {
				counts[discoveryRootOwner(v, uint64(1)<<uint(o.Spec.FieldDegree), h.N-1, w.Size)]++
			}
			for rank, count := range counts {
				starts[rank+1] = starts[rank] + count
			}
			cursor := append([]uint64(nil), starts[:w.Size]...)
			for _, v := range raw[:take] {
				owner := discoveryRootOwner(v, uint64(1)<<uint(o.Spec.FieldDegree), h.N-1, w.Size)
				routed[cursor[owner]] = v
				cursor[owner]++
			}
		}
		gathered := make([]uint64, 0, w.Size)
		for page := 0; page*8 < w.Size; page++ {
			var own [64]byte
			if w.Rank == 0 {
				own = archiveCountPage(counts, page)
			}
			frames, e := w.ControlFrames(own)
			if e != nil {
				return result, e
			}
			if len(frames) != w.Size || frames[w.Rank] != own {
				return result, fmt.Errorf("archive count self agreement")
			}
			values, e := readArchiveCountPage(frames, w.Size, page, take)
			if e != nil {
				return result, e
			}
			gathered = append(gathered, values...)
		}
		total := uint64(0)
		for _, count := range gathered {
			total += count
		}
		if total != take {
			return result, fmt.Errorf("archive count total")
		}
		appendRoots := func(values []uint64) error {
			for _, v := range values {
				if v >= h.N-1 || discoveryRootOwner(v, uint64(1)<<uint(o.Spec.FieldDegree), h.N-1, w.Size) != w.Rank || uint64(len(roots)) >= o.Discovery.MaxRootWords || (len(roots) > 0 && v <= uint64(roots[len(roots)-1])) {
					return fmt.Errorf("cached root ownership/order/admission")
				}
				roots = append(roots, int(v))
			}
			return nil
		}
		for target, count := range gathered {
			if count == 0 {
				continue
			}
			if target == 0 {
				if w.Rank == 0 {
					if err = appendRoots(routed[:count]); err != nil {
						return result, err
					}
				}
				continue
			}
			if w.Rank != 0 && w.Rank != target {
				continue
			}
			values := raw[:count]
			if w.Rank == 0 {
				values = routed[starts[target]:starts[target+1]]
			}
			if err = discoveryWire(w, 4, offset/uint64(rankArchiveTile)+1, h.N, offset, 0, target, values, emit); err != nil {
				return result, err
			}
			if w.Rank == target {
				if err = appendRoots(values); err != nil {
					return result, err
				}
			}
		}
		offset += take
		if offset == take || offset == h.K || time.Since(last) >= 30*time.Second {
			last = time.Now()
			emit(map[string]any{"kind": "workflow-recovery-progress", "case": r.Label, "phase": "cached-discovery", "completed": offset, "total": h.K, "unit": "exponents"})
		}
	}
	var own [64]byte
	if w.Rank == 0 {
		var extra [1]byte
		if n, e := file.Read(extra[:]); n != 0 || e != io.EOF {
			return result, fmt.Errorf("cached archive trailing data")
		}
		observed := fmt.Sprintf("%x", hasher.Sum(nil))
		if observed != source.SupportSHA {
			return result, fmt.Errorf("cached discovery support checksum")
		}
		digest, _ := expectedDigest(observed)
		copy(own[32:], digest)
		if err = file.Close(); err != nil {
			return result, err
		}
	}
	tag, _ := expectedDigest(r.SourceSHA)
	copy(own[:32], tag)
	frames, err := w.ControlFrames(own)
	if err != nil {
		return result, err
	}
	if len(frames) != w.Size || frames[w.Rank] != own {
		return result, fmt.Errorf("cached discovery digest ownership")
	}
	for rank, frame := range frames {
		expected := own
		clear(expected[32:])
		if rank == 0 {
			digest, _ := expectedDigest(source.SupportSHA)
			copy(expected[32:], digest)
		}
		if frame != expected {
			return result, fmt.Errorf("cached discovery source/digest disagreement")
		}
	}
	counts, err := w.Counts(uint64(len(roots)))
	if err != nil {
		return result, err
	}
	var total uint64
	for _, count := range counts {
		total += count
	}
	if total != h.K {
		return result, fmt.Errorf("cached discovery root total")
	}
	passes, hashed := 0, uint64(0)
	if w.Rank == 0 {
		passes, hashed = 1, source.SupportBytes
	}
	emit(map[string]any{"kind": "cached-discovery", "case": r.Label, "source": r.Directory, "source_sha256": r.SourceSHA, "hash_passes": passes, "hashed_bytes": hashed, "local_roots": len(roots), "verified": true})
	return discoveryResult{Roots: roots, K: int(h.K), Samples: source.Samples}, nil
}
