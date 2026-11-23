//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"io"
	"sort"
)

// Options are explicit server-job policy. Random must be provided on rank0;
// production callers choose a suitable secure source,tests choose reproducibility.
// Limits bound individual logical lists/shards,not aggregate process memory.
type rankDiscoveryOptions struct {
	Library, ResultMode                                                     string
	Device                                                                  int
	DeviceBytes, MaxLocalWords, MaxGoodWords, MaxRootWords, MaxArchiveWords uint64
	Random                                                                  io.Reader
	// Optional job admission after exact K is known,before GPU discovery allocations.
	AdmitCode func(ScalingCodeHeader) error
	AdmitHost func(rankHostInventory) error
}

func admitRankDiscovery[Word transposeWord](w *mpi.World, spec Spec, o rankDiscoveryOptions, emit func(map[string]any)) error {
	if w == nil || w.Size < 2 || w.Size > 64 || w.Size&(w.Size-1) != 0 || w.Rank < 0 || w.Rank >= w.Size || spec.FieldDegree < 2 || spec.FieldDegree > 18 || spec.Dimension != 2 || spec.CurveDegree < 1 || emit == nil || (w.Rank == 0 && o.Random == nil) || o.Device < 0 || o.DeviceBytes == 0 {
		return fmt.Errorf("discovery runtime admission")
	}
	if _, err := discoveryResultMode(o.ResultMode); err != nil {
		return err
	}
	var zero Word
	_, wide := any(zero).(uint64)
	if wide != (spec.FieldDegree > 16) {
		return fmt.Errorf("discovery slot width")
	}
	n := uint64(1) << uint(2*spec.FieldDegree)
	_, words := rankBand(n, w.Rank, w.Size)
	if n > uint64(^uint(0)>>1) || words == 0 || words > o.MaxLocalWords || o.MaxGoodWords == 0 || o.MaxRootWords == 0 || o.MaxGoodWords > n || o.MaxRootWords > n || o.MaxArchiveWords > n {
		return fmt.Errorf("discovery host-list admission")
	}
	return nil
}

type discoveryResult struct {
	Roots, Archive []int
	K, Samples     int
	Entropy        string
}

// Counts use integer control frames, not floating-point reductions. Phase and
// sequence bind each exchange; all ranks make the same sample/stop decision.
func discoveryCounts(w *mpi.World, phase, sequence, own, limit uint64) ([]uint64, error) {
	if w == nil || w.Size < 2 || w.Size > 64 || w.Rank < 0 || w.Rank >= w.Size || phase < 1 || phase > 4 || own > limit {
		return nil, fmt.Errorf("discovery count ownership/phase")
	}
	var frame [64]byte
	for j, v := range []uint64{own, phase, sequence, uint64(w.Rank)} {
		binary.LittleEndian.PutUint64(frame[8*j:], v)
	}
	frames, err := w.ControlFrames(frame)
	if err != nil {
		return nil, err
	}
	return validateDiscoveryCounts(frames, w.Size, phase, sequence, limit)
}

func validateDiscoveryCounts(frames [][64]byte, ranks int, phase, sequence, limit uint64) ([]uint64, error) {
	if ranks < 2 || ranks > 64 || len(frames) != ranks {
		return nil, fmt.Errorf("discovery count frame extent")
	}
	counts := make([]uint64, ranks)
	for rank, remote := range frames {
		if binary.LittleEndian.Uint64(remote[8:]) != phase || binary.LittleEndian.Uint64(remote[16:]) != sequence || binary.LittleEndian.Uint64(remote[24:]) != uint64(rank) || !bytes.Equal(remote[32:], make([]byte, 32)) {
			return nil, fmt.Errorf("discovery count phase")
		}
		counts[rank] = binary.LittleEndian.Uint64(remote[:])
	}
	var total uint64
	for _, count := range counts {
		if count > limit-total {
			return nil, fmt.Errorf("discovery count bound")
		}
		total += count
	}
	return counts, nil
}

func discoveryTotal(counts []uint64) uint64 {
	var total uint64
	for _, count := range counts {
		total += count
	}
	return total
}

// Discovery control/data use the existing MPI codecs and before/after proofs.
// kind1=sample coefficients,2=closed-root routing,3=diagnostic support archive,4=cached discovery roots.
func discoveryWire[Word transposeWord](w *mpi.World, kind, sequence, domain, offset uint64, source, target int, data []Word, emit func(map[string]any)) error {
	if w == nil || w.Size < 2 || w.Size > 64 || source < 0 || target < 0 || source >= w.Size || target >= w.Size || source == target || (w.Rank != source && w.Rank != target) || kind < 1 || kind > 4 || domain == 0 || offset > domain || uint64(len(data)) > domain-offset || len(data) == 0 || len(data) > 1<<20 || emit == nil {
		return fmt.Errorf("discovery wire admission")
	}
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
		return fmt.Errorf("discovery packet descriptor: %v", err)
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
		return err
	}
	if err = w.CheckPayloadProof(peer, before, mpi.PayloadFingerprint(data), func(own, remote [64]byte) {
		emit(map[string]any{"kind": "discovery-wire", "phase_kind": kind, "sequence": sequence, "domain": domain, "offset": offset, "words": len(data), "source": source, "target": target, "width": width, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}); err != nil {
		return err
	}
	return nil
}

// Callers must abort all peers on any returned error. No partial result may be
// published. Grids, monomial lists and union bits are rank-local. Random bytes are produced
// in original global-good order by rank0 and sent as bounded coefficient tiles.
// The optional archive sink streams canonical support; nil retains the small
// diagnostic gather. Root ownership is routed
// independently and never depends on that archive or a global support copy.
func discoverRankSupport[Word transposeWord](w *mpi.World, spec Spec, options rankDiscoveryOptions, emit func(map[string]any), archiveSink func([]uint64) error) (result discoveryResult, err error) {
	if err = admitRankDiscovery[Word](w, spec, options, emit); err != nil {
		return result, err
	}
	outputMode, modeErr := discoveryResultMode(options.ResultMode)
	if modeErr != nil {
		return result, modeErr
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
			return result, err
		}
		for j, v := range mask {
			if v != 0 {
				if uint64(len(good)) >= options.MaxGoodWords {
					return result, fmt.Errorf("discovery good-list admission")
				}
				good = append(good, int(first+offset+uint64(j)))
			}
		}
		offset += take
		observe(map[string]any{"kind": "progress", "phase": "local-good-monomials", "completed": offset, "total": words, "unit": "positions"})
	}
	counts, countErr := discoveryCounts(w, 1, 0, uint64(len(good)), n)

	if countErr != nil {
		return result, countErr
	}
	k := int(discoveryTotal(counts))
	h, _, err := canonicalScalingCodeHeader(s, spec.CurveDegree, uint64(k), n)
	if err != nil {
		return result, err
	}
	if options.AdmitCode != nil {
		if err = options.AdmitCode(h); err != nil {
			return result, err
		}
	}
	if options.AdmitHost != nil {
		resultWords := words
		if outputMode == "stream" {
			resultWords = min(words, uint64(4096))
		}
		// Raw support is bounded by the owned band;route/archive lists may grow
		// to twice their admitted length before Go's next append allocation.
		indices := uint64(cap(good)) + min(uint64(k), words) + 2*options.MaxRootWords
		if outputMode == "resident" {
			indices += 2 * options.MaxArchiveWords
		}
		p := hostArrayPhase("discovery", s, n, words, indices, (words+63)/64, resultWords)
		if err = options.AdmitHost(p); err != nil {
			return result, err
		}
	}
	base, err := NewBinaryField(s, h.Polynomial)
	if err != nil {
		return result, err
	}
	tables := newPreparationProducts(base)
	session, err := newRankGPUSession[Word](w, cuda.PreparationOptions{Path: options.Library, S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{options.Device}, MaxElements: n, DeviceBytes: options.DeviceBytes}, 2*s, label, observe)
	if err != nil {
		return result, err
	}
	defer func() {
		if closeErr := session.Close(); err == nil && closeErr != nil {
			err = closeErr
			result = discoveryResult{}
		}
	}()
	executor := &session.Executor
	gpu := executor.GPU
	entropy := sha256.New()
	reader := io.TeeReader(options.Random, entropy)
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
						return result, err
					}
					for j := range tile {
						tile[j] = discoverySampleValue(randomBytes[j*size:(j+1)*size], s)
					}
				}
				if owner != 0 {
					if wireErr := discoveryWire(w, 1, sequence, n, offset, 0, owner, tile, observe); wireErr != nil {
						return result, wireErr
					}
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
		downloadedWords := uint64(0)
		if outputMode == "stream" {
			union := discoveryUnion[Word]{Found: found, Words: words, Limit: n}
			if _, err := session.Discovery(fmt.Sprintf("discovery-%d", samples+1), grid, union.Consume); err != nil {
				return result, err
			}
			added, err := union.Complete()
			if err != nil {
				return result, err
			}
			localCount += added
			downloadedWords = union.Next
		} else {
			values, err := session.Discovery(fmt.Sprintf("discovery-%d", samples+1), grid, nil)
			if err != nil {
				return result, err
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
		totals, countErr := discoveryCounts(w, 2, sequence, localCount, n)

		if countErr != nil {
			return result, countErr
		}
		global = discoveryTotal(totals)
		samples++
		observe(map[string]any{"kind": "discovery-sample", "result_mode": outputMode, "result_resident_host_words": residentWords, "result_buffer_words": resultWords, "result_downloaded_words": downloadedWords, "s": s, "sample": samples, "operation_start": operation, "local_count": localCount, "global_count": global, "k": k, "grid_words": words, "first": first, "good_counts": counts})
	}
	if global != uint64(k) {
		return result, fmt.Errorf("distributed discovery incomplete: %d/%d", global, k)
	}
	raw := make([]int, 0, localCount)
	for j := uint64(0); j < words; j++ {
		if found.has(int(j)) {
			if first+j >= n-1 {
				return result, fmt.Errorf("discovery non-reduced exponent")
			}
			raw = append(raw, int(first+j))
		}
	}
	// Route each closed orbit to its deterministic owner. Count exchange and
	// payload ordering use only local source lists; no global support is needed.
	var roots []int
	for _, index := range raw {
		if discoveryRootOwner(uint64(index), q, n-1, w.Size) == w.Rank {
			if uint64(len(roots)) >= options.MaxRootWords {
				return result, fmt.Errorf("discovery root-list admission")
			}
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
		routed, countErr := discoveryCounts(w, 3, uint64(target), outgoing, uint64(k))

		if countErr != nil {
			return result, countErr
		}
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
						if wireErr := discoveryWire(w, 2, 0, n, sent, source, target, tile, observe); wireErr != nil {
							return result, wireErr
						}
						sent += uint64(at)
						at = 0
					}
				}
				if at > 0 {
					if wireErr := discoveryWire(w, 2, 0, n, sent, source, target, tile[:at], observe); wireErr != nil {
						return result, wireErr
					}
					sent += uint64(at)
				}
				if sent != routed[source] {
					return result, fmt.Errorf("orbit route count")
				}
			} else {
				for received := uint64(0); received < routed[source]; {
					take := min(uint64(len(tile)), routed[source]-received)
					if wireErr := discoveryWire(w, 2, 0, n, received, source, target, tile[:take], observe); wireErr != nil {
						return result, wireErr
					}
					for _, index := range tile[:take] {
						if index >= n-1 || discoveryRootOwner(index, q, n-1, w.Size) != w.Rank {
							return result, fmt.Errorf("orbit route ownership")
						}
						if uint64(len(roots)) >= options.MaxRootWords {
							return result, fmt.Errorf("discovery root-list admission")
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
			return result, err
		}
	}
	totals, countErr := discoveryCounts(w, 4, 0, uint64(len(raw)), uint64(k))

	if countErr != nil {
		return result, countErr
	}
	archive := []int(nil)
	if w.Rank == 0 {
		if archiveSink == nil {
			if uint64(len(raw)) > options.MaxArchiveWords {
				return result, fmt.Errorf("discovery archive admission")
			}
			archive = append(archive, raw...)
		} else {
			for offset := 0; offset < len(raw); {
				take := min(len(tile), len(raw)-offset)
				for j := 0; j < take; j++ {
					tile[j] = uint64(raw[offset+j])
				}
				if err = archiveSink(tile[:take]); err != nil {
					return result, err
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
			if wireErr := discoveryWire(w, 3, 0, n, offset, source, 0, tile[:take], observe); wireErr != nil {
				return result, wireErr
			}
			if w.Rank == 0 {
				if archiveSink != nil {
					if err = archiveSink(tile[:take]); err != nil {
						return result, err
					}
				} else {
					for _, index := range tile[:take] {
						if uint64(len(archive)) >= options.MaxArchiveWords {
							return result, fmt.Errorf("discovery archive admission")
						}
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
	return discoveryResult{roots, archive, k, samples, sum}, nil
}
