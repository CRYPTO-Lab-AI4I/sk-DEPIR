//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

func pilotReference(t *testing.T, s int) (*CodeSupport, *PreparedCode) {
	return pilotReferenceReader(t, s, rand.New(rand.NewSource(109)))
}

func pilotReferenceReader(t *testing.T, s int, random io.Reader) (*CodeSupport, *PreparedCode) {
	t.Helper()
	session, err := NewCodePreprocessor(Spec{s, 2, 2}, preprocessingOptions(s, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	support, err := session.Discover(OrbitOptions{Random: random})
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := session.Prepare(support, PreparationReuse{})
	if err != nil {
		t.Fatal(err)
	}
	return support, prepared
}

// Native-only preflight covers actual K, degree and Newton bounds for the two
// genuine pilot shapes. It initializes neither CUDA nor MPI.
func TestDistributedPilotShapes(t *testing.T) {
	for _, s := range []int{4, 6} {
		support, prepared := pilotReference(t, s)
		k := len(support.exponents)
		n := 1 << uint(2*s)
		if k < 2 || k > n/2 || len(prepared.generator) != n-k {
			t.Fatal("pilot parameters")
		}
	}
}

// This is an assembled two-rank pilot, not a fullsize production entry point.
// Discovery can use coordinator or partitioned grids; publication uses a
// coordinator. P/G stay sharded throughout root/Newton/identity/spectrum;
// explicit stream mode also bounds artifact/support publication buffers.
func TestDistributedPreprocessingPilot(t *testing.T) {
	directory := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if directory == "" {
		t.Skip("explicit complete preprocessing pilot")
	}
	cases := []int{4, 6}
	if os.Getenv("PIR_PIPELINE_SMALL") == "1" {
		cases = cases[:1]
	}
	withDistributedGPUStages(t, directory, func(w *mpi.World, emit func(map[string]any)) {
		for _, s := range cases {
			runDistributedPilot(t, w, directory, s, emit)
		}
		emit(map[string]any{"kind": "completed", "cases": len(cases)})
	})
}

func runDistributedPilot(t *testing.T, w *mpi.World, directory string, s int, emit func(map[string]any)) {
	t.Helper()
	began := time.Now()
	n := 1 << uint(2*s)
	q := uint64(1) << uint(s)
	label := fmt.Sprintf("s%d", s)
	save := func(name string, v []uint32) {
		saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("%s-rank-%d-%s.bin", label, w.Rank, name)), v)
	}
	event := func(phase, state string) {
		emit(map[string]any{"kind": "stage", "case": label, "phase": phase, "event": state})
	}
	resumeFrom := os.Getenv("PIR_PIPELINE_RESUME_FROM")
	publication := os.Getenv("PIR_PIPELINE_PUBLICATION")
	if publication == "" {
		publication = "gather"
	}
	if publication != "gather" && publication != "stream" {
		t.Fatal("publication mode")
	}
	var restoredRoot, restoredPrefix rankPrefixState
	if resumeFrom != "" {
		var e error
		restoredRoot, e = loadRankPrefix(resumeFrom, label, s, w.Rank, w.Size, "root", uint64(n/2))
		if e != nil {
			t.Fatal("resume root", e)
		}
		restoredPrefix, e = loadRankPrefix(resumeFrom, label, s, w.Rank, w.Size, "newton-8", uint64(n/2))
		if e != nil {
			t.Fatal("resume prefix", e)
		}
		if restoredRoot.Record.K != restoredPrefix.Record.K || restoredRoot.Record.SupportSHA != restoredPrefix.Record.SupportSHA {
			t.Fatal("resume cross-record binding")
		}
	}
	event("discovery-reference", "started")
	var err error
	var support, expected *CodeSupport
	var reference *PreparedCode
	layout := os.Getenv("PIR_PIPELINE_DISCOVERY")
	if layout == "" {
		layout = "coordinator"
	}
	if layout != "coordinator" && layout != "distributed" {
		t.Fatal("discovery layout")
	}
	referenceFrom := os.Getenv("PIR_PIPELINE_REFERENCE_FROM")
	var retained pilotReferenceRecord
	if referenceFrom != "" {
		if layout != "distributed" || publication != "stream" {
			t.Fatal("retained reference requires distributed/stream mode")
		}
		var e error
		retained, e = readPilotReference(referenceFrom, s)
		if e != nil {
			t.Fatal(e)
		}
		if w.Rank == 0 {
			if e = retainPilotReference(referenceFrom, directory, retained); e != nil {
				t.Fatal(e)
			}
		}
	}
	expectedEntropy := ""
	if w.Rank == 0 && referenceFrom == "" {
		hash := sha256.New()
		expected, reference = pilotReferenceReader(t, s, io.TeeReader(rand.New(rand.NewSource(109)), hash))
		expectedEntropy = fmt.Sprintf("%x", hash.Sum(nil))
		if resumeFrom != "" {
			support = expected
		} else if layout == "coordinator" {
			options := preprocessingOptions(s, 1)
			options.Backend = "cuda"
			options.GPUPath = os.Getenv("PIR_GPU_PREPARATION_LIBRARY")
			options.GPUDevices = []int{0}
			options.GPUBytes = 64 << 20
			options.Progress = func(p PreprocessingProgress) {
				emit(map[string]any{"kind": "progress", "case": label, "phase": "discovery/" + p.Phase, "event": p.Event, "completed": p.Completed, "total": p.Total, "unit": p.Unit})
			}
			session, err := NewCodePreprocessor(Spec{s, 2, 2}, options)
			if err != nil {
				t.Fatal(err)
			}
			support, err = session.Discover(OrbitOptions{Random: rand.New(rand.NewSource(109))})
			closeErr := session.Close()
			if err != nil || closeErr != nil || !reflect.DeepEqual(support.exponents, expected.exponents) || support.parameters != expected.parameters {
				t.Fatal("GPU/native discovery", err, closeErr)
			}

		}
		save("reference-p", reference.code.p)
		save("reference-g", reference.generator)
		save("reference-spectrum", reference.code.spectrum32)
	}
	var discovered discoveryResult
	streamedSHA := ""
	var archiveFile *os.File
	archiveHash := sha256.New()
	var archive *orderedSupportWriter
	var archiveSink func([]uint64) error
	if referenceFrom != "" && resumeFrom == "" && w.Rank == 0 {
		archiveFile, err = os.OpenFile(filepath.Join(directory, label+"-support.bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer archiveFile.Close()
		archive = &orderedSupportWriter{Writer: io.MultiWriter(archiveFile, archiveHash), S: s, N: uint64(n)}
		archiveSink = archive.append
	}
	if layout == "distributed" && resumeFrom == "" {
		discovered = partitionedDiscovery[uint32](t, w, Spec{s, 2, 2}, emit, archiveSink)
		if w.Rank == 0 {
			expectedSamples := 0
			if referenceFrom == "" {
				if !reflect.DeepEqual(discovered.Archive, expected.exponents) {
					t.Fatal("partitioned discovery/native support mismatch")
				}
				expectedSamples = expected.parameters.Samples
				support = &CodeSupport{parameters: expected.parameters, exponents: discovered.Archive}
			} else {
				expectedEntropy = retained.Entropy
				expectedSamples = retained.Parameters.Samples
				if len(discovered.Archive) != 0 || archive.Count != retained.Parameters.K || discovered.K != int(retained.Parameters.K) {
					t.Fatal("streamed support count")
				}
				if err = archiveFile.Sync(); err != nil {
					t.Fatal(err)
				}
				if err = archiveFile.Close(); err != nil {
					t.Fatal(err)
				}
				streamedSHA = fmt.Sprintf("%x", archiveHash.Sum(nil))
				if streamedSHA != retained.SHA[label+"-support.bin"] {
					t.Fatal("streamed support checksum")
				}
			}
			if discovered.Samples != expectedSamples || discovered.Entropy != expectedEntropy {
				t.Fatal("partitioned discovery/native sampling mismatch")
			}
			emit(map[string]any{"kind": "discovery-reference", "discovery": true, "case": label, "s": s, "samples": discovered.Samples, "entropy_sha256": discovered.Entropy, "reference_entropy_sha256": expectedEntropy, "matches": true})
		}
	}
	count := 0
	if referenceFrom != "" {
		count = int(retained.Parameters.K)
	} else if support != nil {
		count = len(support.exponents)
	}
	k := int(w.ShareCount(uint64(count)))
	if k < 2 || k > n/2 {
		t.Fatal("pilot K")
	}
	h, field, err := canonicalScalingCodeHeader(s, 2, uint64(k), uint64(n))
	if err != nil {
		t.Fatal(err)
	}
	if resumeFrom != "" && restoredRoot.Record.K != k {
		t.Fatal("resume/native K binding")
	}
	base, err := NewBinaryField(s, h.Polynomial)
	if err != nil {
		t.Fatal(err)
	}
	tables := newPreparationProducts(base)
	var binding [64]byte
	if w.Rank == 0 {
		digest := [32]byte{}
		if referenceFrom != "" {
			if resumeFrom != "" {
				if err = copyPilotSupport(referenceFrom, directory, label, retained); err != nil {
					t.Fatal(err)
				}
				streamedSHA = retained.SHA[label+"-support.bin"]
			}
			raw, e := hex.DecodeString(streamedSHA)
			if e != nil || len(raw) != 32 {
				t.Fatal("support digest", e)
			}
			copy(digest[:], raw)
		} else {
			var data bytes.Buffer
			if _, err = support.Write(&data); err != nil {
				t.Fatal(err)
			}
			digest = sha256.Sum256(data.Bytes())
			if err = os.WriteFile(filepath.Join(directory, label+"-support.bin"), data.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
		copy(binding[:32], digest[:])
		copy(binding[32:], digest[:])
	}

	remote, err := w.ExchangeControl(1-w.Rank, binding)
	if err != nil {
		t.Fatal(err)
	}
	if w.Rank == 1 {
		binding = remote
	}
	if !bytes.Equal(binding[:32], binding[32:]) {
		t.Fatal("support binding")
	}
	supportSHA := fmt.Sprintf("%x", binding[:32])
	if resumeFrom != "" && supportSHA != restoredRoot.Record.SupportSHA {
		t.Fatal("resume/native support binding")
	}
	event("discovery-reference", "completed")
	event("route-support", "started")
	var shard []int
	proofSupport := func(direction string, values []uint64, before [32]byte) {
		if err := w.CheckPayloadProof(1-w.Rank, before, mpi.PayloadFingerprint(values), func(own, remote [64]byte) {
			emit(map[string]any{"kind": "support-proof", "case": label, "direction": direction, "words": len(values), "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
		}); err != nil {
			t.Fatal(err)
		}
	}
	if resumeFrom != "" {
		data, e := readBoundedPilotFile(filepath.Join(resumeFrom, fmt.Sprintf("%s-rank-%d-support.bin", label, w.Rank)), 8*uint64(k))
		if e != nil || len(data)%8 != 0 {
			t.Fatal("resume support shard", e)
		}
		for offset := 0; offset < len(data); offset += 8 {
			index := binary.LittleEndian.Uint64(data[offset:])
			if index >= uint64(n-1) {
				t.Fatal("resume support range")
			}
			shard = append(shard, int(index))
		}
		if len(shard) > 0 {
			if e := validateExponentSupportWith(supportSorted, shard, n, q, nil); e != nil {
				t.Fatal(e)
			}
		}
	} else if layout == "distributed" {
		shard = discovered.Roots
	} else {
		if w.Rank == 0 {
			err = routeRootSupport(supportSorted, support.exponents, n, q, 2, 127, func(owner int, values []uint64) error {
				if owner == 0 {
					for _, v := range values {
						shard = append(shard, int(v))
					}
					return nil
				}
				before := mpi.PayloadFingerprint(values)
				if err := w.SendIndices(1, values); err != nil {
					return err
				}
				proofSupport("send", values, before)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if err = w.SendIndices(1, nil); err != nil {
				t.Fatal(err)
			}
		} else {
			for {
				values, e := w.ReceiveIndices()
				if e != nil {
					t.Fatal(e)
				}
				if len(values) == 0 {
					break
				}
				proofSupport("receive", values, mpi.PayloadFingerprint(values))
				for _, v := range values {
					shard = append(shard, int(v))
				}
			}
		}
	}

	sort.Ints(shard)
	leftCount := 0
	if w.Rank == 0 {
		leftCount = len(shard) + 1
	}
	ac := w.ShareCount(uint64(leftCount))
	bc := uint64(k+2) - ac
	if uint64(len(shard)+1) != ([2]uint64{ac, bc})[w.Rank] {
		t.Fatal("support ownership count")
	}
	indices := make([]uint64, len(shard))
	for j, v := range shard {
		indices[j] = uint64(v)
	}
	saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("%s-rank-%d-support.bin", label, w.Rank)), indices)
	event("route-support", "completed")
	if resumeFrom == "" {
		event("local-roots", "started")
	} else {
		event("local-roots", "restored")
	}
	options := cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: uint64(n), DeviceBytes: 64 << 20}
	fixture := &rootFixture{header: h, extension: field, n: n, maximum: n, base: orbitField{q: q, base: base, lookup: tables}}
	var local []uint32
	if resumeFrom == "" {
		factor, err := cuda.NewPreparation(options)
		if err != nil {
			t.Fatal(err)
		}

		local, err = fixture.root(supportSorted, factor, shard, func(stage string) { event("roots/"+stage, "started") }, func(stage string, done, total uint64, unit string) {
			if done == total {
				emit(map[string]any{"kind": "progress", "case": label, "phase": "roots/" + stage, "completed": done, "total": total, "unit": unit})
			}
		})
		closeErr := factor.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}

	}
	if resumeFrom == "" {
		event("local-roots", "completed")
	}
	gpu, err := cuda.NewRankStorage(options, uint64(1)<<uint(s), uint64(1)<<uint(s), w.Rank, w.Size, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Close()
	p, err := newRankTransposePlan(1, uint64(1)<<uint(s), uint64(1)<<uint(s), w.Rank, w.Size, uint64(n), 4096)
	if err != nil {
		t.Fatal(err)
	}
	capacity := min(uint64(n), 3*(uint64(1)<<uint(s)))
	first := p.InputFirst * p.Columns
	words := p.InputWords
	call := ""
	callOp := uint64(0)
	pass := 0
	span := uint64(2 + 2*s)
	notice := func(phase string, done, total uint64) {
		if done == 0 {
			switch phase {
			case "rank-product-left-forward":
				pass = 0
			case "rank-product-right-forward":
				pass = 1
			case "rank-product-inverse":
				pass = 2
			}
		}
		if done == 0 || done == total {
			emit(map[string]any{"kind": "progress", "case": label, "call": call, "phase": phase, "pass": pass, "completed": done, "total": total})
		}
	}
	executor := rankGPUExecutor[uint32]{GPU: gpu, Plan: p, Capacity: capacity, Limit: q, Notice: notice}
	executor.Transpose = transposeMPI[uint32]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, data []uint32, own, remote [64]byte) {
		emit(map[string]any{"kind": "tile-proof", "case": call, "pass": pass, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	executor.Windows = taylorMPI[uint32]{world: w, record: func(direction string, peer int, p taylorPacket, data []uint32, own, remote [64]byte) {
		emit(map[string]any{"kind": "taylor-proof", "case": call, "pass": pass, "packet": p, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	executor.Degrees = func(level int) ([]uint64, error) {
		degrees, e := gpu.TaylorDegrees(level)
		if e == nil {
			emit(map[string]any{"kind": "taylor-plan", "case": call, "pass": pass, "operation_start": callOp + uint64(pass)*span, "level": level, "degrees": degrees, "capacity": capacity})
		}
		return degrees, e
	}
	product := func(name string, a, b []uint32, aCount, bCount, keep uint64) []uint32 {
		call, callOp, pass = label+"/"+name, executor.Plan.Operation, 0
		result, e := executor.Product(a, b, aCount, bCount, keep)
		if e != nil {
			t.Fatal(name, e)
		}
		emit(map[string]any{"kind": "executor-call", "case": call, "s": s, "operation_start": callOp, "mode": "product", "a_count": aCount, "b_count": bCount, "keep": keep, "matches": true})
		return result
	}
	link := operandMPI{world: w, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
		emit(map[string]any{"kind": "operand-proof", "case": label, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	var root []uint32
	if resumeFrom != "" {
		root = restoredRoot.Values
	} else {
		var inputs [2][]uint32
		inputs[w.Rank] = local
		pair, err := routePolynomialOperands(rankOperandPlan{Operation: 1, Domain: uint64(n), TileWords: 4096, Limit: q, MaxLocalWords: words, Counts: [2]uint64{ac, bc}, Owners: [2]int{0, 1}, Rank: w.Rank, Ranks: w.Size}, inputs, link, nil)
		if err != nil {
			t.Fatal(err)
		}
		root = product("root", pair[0], pair[1], ac, bc, uint64(k+1))

	}
	save("p", root)
	// Each completed prefix is durably saved on every rank, then committed before
	// reuse. Reload one interior prefix in-process to exercise the resume boundary;
	// this does not claim recovery from a fresh process or failed node.
	persist := func(stage string, values []uint32, keep int, reload bool) []uint32 {
		save(stage, values)
		own := mpi.PayloadFingerprint(values)
		var frame [64]byte
		copy(frame[:32], own[:])
		copy(frame[32:], own[:])
		other, e := w.ExchangeControl(1-w.Rank, frame)
		if e != nil || !bytes.Equal(other[:32], other[32:]) {
			t.Fatal("checkpoint peer digest", e)
		}
		w.Barrier()
		manifest := map[string]any{"s": s, "k": k, "n": n, "prefix": keep, "ranks": 2, "support_sha256": supportSHA, "stage": stage}
		hashes := []string{fmt.Sprintf("%x", own), fmt.Sprintf("%x", other[:32])}
		if w.Rank == 1 {
			hashes[0], hashes[1] = hashes[1], hashes[0]
		}
		manifest["sha256"] = hashes
		if w.Rank == 0 {
			data, _ := json.MarshalIndent(manifest, "", "  ")
			file, e := os.OpenFile(filepath.Join(directory, label+"-"+stage+"-commit.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = file.Write(data); e == nil {
				e = file.Sync()
			}
			ce := file.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
		}
		w.Barrier()
		if reload {
			expectedManifest, _ := json.MarshalIndent(manifest, "", "  ")
			commit, e := os.ReadFile(filepath.Join(directory, label+"-"+stage+"-commit.json"))
			if e != nil || !bytes.Equal(commit, expectedManifest) {
				t.Fatal("checkpoint commit binding", e)
			}
			data, e := os.ReadFile(filepath.Join(directory, fmt.Sprintf("%s-rank-%d-%s.bin", label, w.Rank, stage)))
			if e != nil || len(data) != 4*len(values) || sha256.Sum256(data) != own {
				t.Fatal("checkpoint reload", e)
			}
			loaded := make([]uint32, len(values))
			for j := range loaded {
				loaded[j] = binary.LittleEndian.Uint32(data[4*j:])
			}
			values = loaded
		}
		emit(map[string]any{"kind": "checkpoint", "case": label, "stage": stage, "prefix": keep, "reloaded": reload, "support_sha256": supportSHA})
		return values
	}
	root = persist("root", root, k+1, false)
	g := make([]uint32, words)
	known := 1
	reloaded := false
	if resumeFrom != "" {
		g = restoredPrefix.Values
		known = restoredPrefix.Record.Prefix
		reloaded = true
		g = persist(fmt.Sprintf("newton-%d", known), g, known, false)
		emit(map[string]any{"kind": "resume", "case": label, "prefix": known, "root_support_sha256": restoredRoot.Record.SupportSHA, "source": resumeFrom, "pid": os.Getpid(), "verified": true})
	} else if first == 0 {
		g[0] = 1
	}
	initialPrefix := known
	err = reciprocalSchedule(known, n-k, func(size int) error {
		squareOp := executor.Plan.Operation
		executor.Plan.Operation++
		squared, e := squarePolynomialShard(rankSquarePlan{squareOp, uint64(n), 4096, q, uint64(size), words, w.Rank, w.Size}, g, func(v uint32) uint32 { return fixture.base.mul(v, v) }, link, func(done, total uint64) { notice("newton-square", done, total) })
		if e != nil {
			return e
		}
		aCount := min(k+1, size)
		a := append([]uint32(nil), root...)
		for j := range a {
			if first+uint64(j) >= uint64(aCount) {
				a[j] = 0
			}
		}
		squareCount := min(2*known-1, size)
		g = product(fmt.Sprintf("newton-%d", size), a, squared, uint64(aCount), uint64(squareCount), uint64(size))
		reload := !reloaded && size >= 8 && size < n-k
		g = persist(fmt.Sprintf("newton-%d", size), g, size, reload)
		reloaded = reloaded || reload
		known = size
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded {
		t.Fatal("missing interior prefix reload")
	}
	identity := product("identity", root, g, uint64(k+1), uint64(n-k), uint64(n))
	for j, v := range identity {
		var want uint32
		if index := first + uint64(j); index == 0 || index == uint64(n-1) {
			want = 1
		}
		if v != want {
			t.Fatal("full P*G identity")
		}
	}
	save("g", g)
	save("identity", identity)
	call, callOp, pass = label+"/spectrum", executor.Plan.Operation, 0
	spectrum, err := executor.Spectrum(g, uint64(n-k))
	if err != nil {
		t.Fatal(err)
	}
	save("spectrum", spectrum)
	emit(map[string]any{"kind": "executor-call", "case": call, "s": s, "operation_start": callOp, "mode": "spectrum", "keep": n, "matches": true})
	event("artifact-publication", "started")
	if publication == "stream" {
		publishPilotStream(t, w, directory, label, h, root, g, spectrum, reference, link, &executor.Plan.Operation)
	} else {
		publish := func(values []uint32) []uint32 {
			op := executor.Plan.Operation
			executor.Plan.Operation++
			packet := operandPacket{op, uint64(n), 0, words, words, 1, 0}
			if w.Rank == 1 {
				if e := link.SendOperand(packet, values); e != nil {
					t.Fatal(e)
				}
				return nil
			}
			result := make([]uint32, n)
			copy(result, values)
			if e := link.ReceiveOperand(packet, result[words:]); e != nil {
				t.Fatal(e)
			}
			return result
		}
		fullP, fullG, fullSpectrum := publish(root), publish(g), publish(spectrum)
		if w.Rank == 0 {
			if !reflect.DeepEqual(fullP[:k+1], reference.code.p) || !reflect.DeepEqual(fullG[:n-k], reference.generator) || !reflect.DeepEqual(fullSpectrum, reference.code.spectrum32) {
				t.Fatal("complete canonical output oracle")
			}
			code := &ScalingCodeData{header: h, field: field, p: fullP[:k+1], spectrum32: fullSpectrum}
			var actual, expected bytes.Buffer
			if err = WriteScalingCode(&actual, code); err != nil {
				t.Fatal(err)
			}
			if err = WriteScalingCode(&expected, reference.code); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual.Bytes(), expected.Bytes()) {
				t.Fatal("canonical artifact bytes")
			}
			for name, data := range map[string][]byte{"code": actual.Bytes(), "reference-code": expected.Bytes()} {
				file, e := os.OpenFile(filepath.Join(directory, label+"-"+name+".bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = file.Write(data); e == nil {
					e = file.Sync()
				}
				ce := file.Close()
				if e != nil || ce != nil {
					t.Fatal(e, ce)
				}
			}
		}

	}
	w.Barrier()
	event("artifact-publication", "completed")
	emit(map[string]any{"kind": "pipeline", "case": label, "s": s, "n": n, "k": k, "first": first, "words": words, "matches": true, "identity_verified": true, "prefix_reloaded": reloaded, "owned_device_bytes": gpu.OwnedBytes(), "seconds": w.Max(time.Since(began).Seconds()), "root_a_count": ac, "root_b_count": bc, "discovery_layout": layout, "initial_prefix": initialPrefix, "discovery_performed": resumeFrom == "", "root_performed": resumeFrom == "", "publication": publication, "pid": os.Getpid(), "reference_mode": map[bool]string{true: "retained", false: "native"}[referenceFrom != ""], "support_streamed": referenceFrom != ""})
}

// Explicit bounded native-reference export for the pilot's independent inputs.
// Setup is kept separate from distributed runtime and retained with checksums.
func TestExportPilotReferences(t *testing.T) {
	directory := os.Getenv("PIR_PIPELINE_REFERENCE_EXPORT")
	if directory == "" {
		t.Skip("explicit bounded native reference export")
	}
	for _, s := range []int{4, 6} {
		entropy := sha256.New()
		support, prepared := pilotReferenceReader(t, s, io.TeeReader(rand.New(rand.NewSource(109)), entropy))
		label := fmt.Sprintf("s%d", s)
		for stage, values := range map[string][]uint32{"p": prepared.code.p, "g": prepared.generator, "spectrum": prepared.code.spectrum32} {
			saveMatrixStageWords(t, filepath.Join(directory, label+"-rank-0-reference-"+stage+".bin"), values)
		}
		write := func(name string, fn func(io.Writer) error) {
			f, e := os.OpenFile(filepath.Join(directory, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e != nil {
				t.Fatal(e)
			}
			e = fn(f)
			if e == nil {
				e = f.Sync()
			}
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
		}
		write(label+"-reference-code.bin", func(w io.Writer) error { return WriteScalingCode(w, prepared.code) })
		write(label+"-support.bin", func(w io.Writer) error { _, e := support.Write(w); return e })
		record := pilotReferenceRecord{Parameters: support.parameters, Entropy: fmt.Sprintf("%x", entropy.Sum(nil)), SHA: map[string]string{}}
		for name := range pilotReferenceSizes(record.Parameters) {
			data, e := os.ReadFile(filepath.Join(directory, name))
			if e != nil {
				t.Fatal(e)
			}
			record.SHA[name] = fmt.Sprintf("%x", sha256.Sum256(data))
		}
		write(label+"-reference.json", func(w io.Writer) error { return json.NewEncoder(w).Encode(record) })
		t.Logf("s=%d reference files exported without GPU/MPI initialization", s)
	}
}
