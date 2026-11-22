//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

type discoveryResult struct {
	Roots, Archive []int
	K, Samples     int
	Entropy        string
}

// Counts use integer control frames, not floating-point reductions. Phase and
// sequence bind each exchange; all ranks make the same sample/stop decision.
func discoveryCounts(t *testing.T, w *mpi.World, phase, sequence, own, limit uint64) []uint64 {
	t.Helper()
	var frame [64]byte
	for j, v := range []uint64{own, phase, sequence, uint64(w.Rank)} {
		binary.LittleEndian.PutUint64(frame[8*j:], v)
	}
	frames, err := w.ControlFrames(frame)
	if err != nil {
		t.Fatal(err)
	}
	counts := make([]uint64, w.Size)
	for rank, remote := range frames {
		if binary.LittleEndian.Uint64(remote[8:]) != phase || binary.LittleEndian.Uint64(remote[16:]) != sequence || binary.LittleEndian.Uint64(remote[24:]) != uint64(rank) || !bytes.Equal(remote[32:], make([]byte, 32)) {
			t.Fatal("discovery count phase")
		}
		counts[rank] = binary.LittleEndian.Uint64(remote[:])
	}
	var total uint64
	for _, count := range counts {
		if count > limit-total {
			t.Fatal("discovery count bound")
		}
		total += count
	}
	return counts
}

func discoveryTotal(counts []uint64) uint64 {
	var total uint64
	for _, count := range counts {
		total += count
	}
	return total
}

// Discovery control/data use the existing MPI codecs and before/after proofs.
// kind1=sample coefficients,2=closed-root routing,3=diagnostic support archive.
func discoveryWire[Word transposeWord](t *testing.T, w *mpi.World, kind, sequence, domain, offset uint64, source, target int, data []Word, emit func(map[string]any)) {
	width := uint64(8)
	var zero Word
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	fields := [8]uint64{kind, sequence, domain, offset, uint64(len(data)), uint64(source), uint64(target), width}
	var frame [64]byte
	for j, v := range fields {
		binary.LittleEndian.PutUint64(frame[8*j:], v)
	}
	peer := source
	if w.Rank == source {
		peer = target
	}
	other, err := w.ExchangeControl(peer, frame)
	if err != nil || other != frame {
		t.Fatal("discovery packet descriptor", err)
	}
	before := mpi.PayloadFingerprint(data)
	direction := "send"
	if w.Rank == source {
		switch v := any(data).(type) {
		case []uint32:
			err = w.SendCoefficients32(peer, v)
		case []uint64:
			err = w.SendCoefficients64(peer, v)
		}
	} else {
		direction = "receive"
		switch v := any(data).(type) {
		case []uint32:
			err = w.ReceiveCoefficients32(peer, v)
		case []uint64:
			err = w.ReceiveCoefficients64(peer, v)
		}
		before = mpi.PayloadFingerprint(data)
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = w.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(data), func(own, remote [64]byte) {
		emit(map[string]any{"kind": "discovery-wire", "phase_kind": kind, "sequence": sequence, "domain": domain, "offset": offset, "words": len(data), "source": source, "target": target, "width": width, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}); err != nil {
		t.Fatal(err)
	}
}

// Grids, monomial lists and union bits are rank-local. Random bytes are produced
// in original global-good order by rank0 and sent as bounded coefficient tiles.
// The optional archive sink streams canonical support; nil retains the small
// diagnostic gather. Root ownership is routed
// independently and never depends on that archive or a global support copy.
func partitionedDiscovery[Word transposeWord](t *testing.T, w *mpi.World, spec Spec, emit func(map[string]any), archiveSink func([]uint64) error) discoveryResult {
	t.Helper()
	if w.Size < 2 || w.Size > 64 || w.Size&(w.Size-1) != 0 {
		t.Fatal("partitioned discovery requires power-of-two ranks through64")
	}
	outputMode, modeErr := discoveryResultMode(os.Getenv("PIR_PIPELINE_DISCOVERY_RESULT"))
	if modeErr != nil {
		t.Fatal(modeErr)
	}
	s := spec.FieldDegree
	n := uint64(1) << uint(2*s)
	q := uint64(1) << uint(s)
	first, words := rankBand(n, w.Rank, w.Size)
	label := fmt.Sprintf("s%d", s)
	observe := func(r map[string]any) {
		r["discovery"] = true
		if _, ok := r["case"]; !ok {
			r["case"] = label
		}
		emit(r)
	}
	var good []int
	for offset := uint64(0); offset < words; {
		take := min(uint64(65536), words-offset)
		mask, err := cpp.GoodMonomialRange(s, spec.CurveDegree, first+offset, int(take), 1)
		if err != nil {
			t.Fatal(err)
		}
		for j, v := range mask {
			if v != 0 {
				good = append(good, int(first+offset+uint64(j)))
			}
		}
		offset += take
		observe(map[string]any{"kind": "progress", "phase": "local-good-monomials", "completed": offset, "total": words, "unit": "positions"})
	}
	counts := discoveryCounts(t, w, 1, 0, uint64(len(good)), n)
	k := int(discoveryTotal(counts))
	h, _, err := canonicalScalingCodeHeader(s, spec.CurveDegree, uint64(k), n)
	if err != nil {
		t.Fatal(err)
	}
	base, err := NewBinaryField(s, h.Polynomial)
	if err != nil {
		t.Fatal(err)
	}
	tables := newPreparationProducts(base)
	gpu, err := cuda.NewRankStorage(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 64 << 20}, q, q, w.Rank, w.Size, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Close()
	plan, err := newRankTransposePlan(1, q, q, w.Rank, w.Size, n, 4096)
	if err != nil {
		t.Fatal(err)
	}
	call := ""
	op := uint64(0)
	capacity := min(n, 3*q)
	executor := rankGPUExecutor[Word]{GPU: gpu, Plan: plan, Capacity: capacity, Limit: q}
	lastNotice := map[string]time.Time{}
	executor.Notice = func(phase string, done, total uint64) {
		key := call + "/" + phase
		if done == 0 || done == total || time.Since(lastNotice[key]) >= 30*time.Second {
			lastNotice[key] = time.Now()
			observe(map[string]any{"kind": "progress", "phase": phase, "completed": done, "total": total, "call": call})
		}
	}
	executor.Transpose = transposeMPI[Word]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, data []Word, own, remote [64]byte) {
		observe(map[string]any{"kind": "tile-proof", "case": call, "pass": 0, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	executor.Windows = taylorMPI[Word]{world: w, record: func(direction string, peer int, p taylorPacket, data []Word, own, remote [64]byte) {
		observe(map[string]any{"kind": "taylor-proof", "case": call, "pass": 0, "packet": p, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	executor.Degrees = func(level int) ([]uint64, error) {
		degrees, e := gpu.TaylorDegrees(level)
		if e == nil {
			observe(map[string]any{"kind": "taylor-plan", "case": call, "pass": 0, "operation_start": op, "level": level, "degrees": degrees, "capacity": capacity})
		}
		return degrees, e
	}
	random := rand.New(rand.NewSource(109))
	entropy := sha256.New()
	reader := io.TeeReader(random, entropy)
	grid := make([]uint32, words)
	found := newExponentSet(int(words))
	sampleTile := make([]uint32, 4096)
	randomBytes := make([]byte, 4096*discoverySampleWidth(s))
	samples := 0
	global := uint64(0)
	localCount := uint64(0)
	totalDownloaded := uint64(0)
	resultWords, residentWords := words, words
	if outputMode == "stream" {
		resultWords, residentWords = min(words, executor.Plan.TileWords), 0
	}
	for samples < discoverySampleLimit(s) && global < uint64(k) {
		clear(grid)
		sequence := uint64(samples + 1)
		for owner := 0; owner < w.Size; owner++ {
			if w.Rank != 0 && w.Rank != owner {
				continue
			}
			for offset := uint64(0); offset < counts[owner]; {
				take := min(uint64(4096), counts[owner]-offset)
				tile := sampleTile[:take]
				if w.Rank == 0 {
					size := discoverySampleWidth(s)
					if _, err = io.ReadFull(reader, randomBytes[:int(take)*size]); err != nil {
						t.Fatal(err)
					}
					for j := range tile {
						tile[j] = discoverySampleValue(randomBytes[j*size:(j+1)*size], s)
					}
				}
				if owner != 0 {
					discoveryWire(t, w, 1, sequence, n, offset, 0, owner, tile, observe)
				}
				if w.Rank == owner {
					for j, v := range tile {
						grid[uint64(good[offset+uint64(j)])-first] = v
					}
				}
				offset += take
			}
		}
		operation := executor.Plan.Operation
		observeTransform := func(inverse bool, start uint64) {
			op = start
			call = fmt.Sprintf("%s/discovery-%d/base", label, samples+1)
			if inverse {
				call = fmt.Sprintf("%s/discovery-%d/inverse", label, samples+1)
			}
		}
		downloadedWords := uint64(0)
		if outputMode == "stream" {
			union := discoveryUnion[Word]{Found: found, Words: words, Limit: n}
			if err := executor.DiscoveryTiles(grid, observeTransform, union.Consume); err != nil {
				t.Fatal(err)
			}
			added, err := union.Complete()
			if err != nil {
				t.Fatal(err)
			}
			localCount += added
			downloadedWords = union.Next
		} else {
			values, err := executor.Discovery(grid, observeTransform)
			if err != nil {
				t.Fatal(err)
			}
			downloadedWords = uint64(len(values))
			for j, value := range values {
				if value != 0 && !found.has(j) {
					found.add(j)
					localCount++
				}
			}
		}
		totalDownloaded += downloadedWords
		totals := discoveryCounts(t, w, 2, sequence, localCount, n)
		global = discoveryTotal(totals)
		samples++
		observe(map[string]any{"kind": "discovery-sample", "result_mode": outputMode, "result_resident_host_words": residentWords, "result_buffer_words": resultWords, "result_downloaded_words": downloadedWords, "s": s, "sample": samples, "operation_start": operation, "local_count": localCount, "global_count": global, "k": k, "grid_words": words, "first": first, "good_counts": counts})
	}
	if global != uint64(k) {
		t.Fatal("distributed discovery incomplete", global, k)
	}
	raw := make([]int, 0, localCount)
	for j := uint64(0); j < words; j++ {
		if found.has(int(j)) {
			if first+j >= n-1 {
				t.Fatal("discovery non-reduced exponent")
			}
			raw = append(raw, int(first+j))
		}
	}
	// Route each closed orbit to its deterministic owner. Count exchange and
	// payload ordering use only local source lists; no global support is needed.
	var roots []int
	for _, index := range raw {
		if discoveryRootOwner(uint64(index), q, n-1, w.Size) == w.Rank {
			roots = append(roots, index)
		}
	}
	tile := make([]uint64, 4096)
	for target := 0; target < w.Size; target++ {
		var outgoing uint64
		for _, index := range raw {
			if target != w.Rank && discoveryRootOwner(uint64(index), q, n-1, w.Size) == target {
				outgoing++
			}
		}
		routed := discoveryCounts(t, w, 3, uint64(target), outgoing, uint64(k))
		for source := 0; source < w.Size; source++ {
			if source == target || (w.Rank != source && w.Rank != target) {
				continue
			}
			if w.Rank == source {
				sent, at := uint64(0), 0
				for _, index := range raw {
					if discoveryRootOwner(uint64(index), q, n-1, w.Size) != target {
						continue
					}
					tile[at] = uint64(index)
					at++
					if at == len(tile) {
						discoveryWire(t, w, 2, 0, n, sent, source, target, tile, observe)
						sent += uint64(at)
						at = 0
					}
				}
				if at > 0 {
					discoveryWire(t, w, 2, 0, n, sent, source, target, tile[:at], observe)
					sent += uint64(at)
				}
				if sent != routed[source] {
					t.Fatal("orbit route count")
				}
			} else {
				for received := uint64(0); received < routed[source]; {
					take := min(uint64(len(tile)), routed[source]-received)
					discoveryWire(t, w, 2, 0, n, received, source, target, tile[:take], observe)
					for _, index := range tile[:take] {
						if index >= n-1 || discoveryRootOwner(index, q, n-1, w.Size) != w.Rank {
							t.Fatal("orbit route ownership")
						}
						roots = append(roots, int(index))
					}
					received += take
				}
			}
		}
	}
	sort.Ints(roots)
	if len(roots) > 0 {
		if err = validateExponentSupportWith(supportSorted, roots, int(n), q, nil); err != nil {
			t.Fatal(err)
		}
	}
	totals := discoveryCounts(t, w, 4, 0, uint64(len(raw)), uint64(k))
	archive := []int(nil)
	if w.Rank == 0 {
		if archiveSink == nil {
			archive = append(archive, raw...)
		} else {
			for offset := 0; offset < len(raw); {
				take := min(len(tile), len(raw)-offset)
				for j := 0; j < take; j++ {
					tile[j] = uint64(raw[offset+j])
				}
				if err = archiveSink(tile[:take]); err != nil {
					t.Fatal(err)
				}
				offset += take
			}
		}
	}
	for source := 1; source < w.Size; source++ {
		if w.Rank != 0 && w.Rank != source {
			continue
		}
		for offset := uint64(0); offset < totals[source]; {
			take := min(uint64(len(tile)), totals[source]-offset)
			if w.Rank == source {
				for j := uint64(0); j < take; j++ {
					tile[j] = uint64(raw[offset+j])
				}
			}
			discoveryWire(t, w, 3, 0, n, offset, source, 0, tile[:take], observe)
			if w.Rank == 0 {
				if archiveSink != nil {
					if err = archiveSink(tile[:take]); err != nil {
						t.Fatal(err)
					}
				} else {
					for _, index := range tile[:take] {
						archive = append(archive, int(index))
					}
				}
			}
			offset += take
		}

	}
	sum := ""
	if w.Rank == 0 {
		sum = fmt.Sprintf("%x", entropy.Sum(nil))
	}
	observe(map[string]any{"kind": "discovery-result", "result_mode": outputMode, "result_resident_host_words": residentWords, "result_buffer_words": resultWords, "result_downloaded_words": totalDownloaded, "s": s, "samples": samples, "k": k, "grid_words": words, "first": first, "local_good": len(good), "raw_count": len(raw), "root_count": len(roots), "entropy_sha256": sum, "owned_device_bytes": gpu.OwnedBytes()})
	return discoveryResult{roots, archive, k, samples, sum}
}

// Bound the actual assembled shapes and validate the new orbit ownership before
// an allocation; no CUDA/MPI initialization and no new calibration workload.
func TestPartitionedDiscoveryShapes(t *testing.T) {
	for _, s := range []int{4, 6} {
		support, _ := pilotReference(t, s)
		n := uint64(1) << uint(2*s)
		q := uint64(1) << uint(s)
		count := 0
		owners := [2][]int{}
		for rank := 0; rank < 2; rank++ {
			first, words := rankBand(n, rank, 2)
			mask, err := cpp.GoodMonomialRange(s, 2, first, int(words), 1)
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range mask {
				count += int(v)
			}
		}
		if count != len(support.exponents) {
			t.Fatal("local monomial count")
		}
		for _, index := range support.exponents {
			owner := discoveryRootOwner(uint64(index), q, n-1, 2)
			owners[owner] = append(owners[owner], index)
		}
		for _, part := range owners {
			if len(part) == 0 {
				t.Fatal("empty pilot orbit owner")
			}
			if err := validateExponentSupportWith(supportSorted, part, int(n), q, nil); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("s=%d N=%d K=%d samples=%d owned roots=%d/%d", s, n, count, support.parameters.Samples, len(owners[0]), len(owners[1]))
	}
}
