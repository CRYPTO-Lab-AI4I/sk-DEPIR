// pir-preprocess is the experimental persistent server preprocessing runner.
// CUDA is selected explicitly at build and run time; CPU client builds stay free
// of CUDA. See docs/preprocessing.md for cache trust and resource admission.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Injected by reproducible experiment builds, alongside source-file hashes in
// the job record. Unknown is honest for manual builds without VCS provenance.
var sourceRevision = "unknown"

type config struct {
	checkpointInterval                                time.Duration
	s, t, workers                                     int
	backend, devices, library, layout, cache, profile string
	arrayBytes, nativeBytes, gpuBytes, codeBytes      uint64
	spectrumTile                                      uint64
	resume, dimensionOnly                             bool
}

func parse(args []string) (config, error) {
	var c config
	f := flag.NewFlagSet("pir-preprocess", flag.ContinueOnError)
	f.IntVar(&c.s, "s", 0, "base field degree (2..17); m=2")
	f.IntVar(&c.t, "t", 5, "curve degree")
	f.IntVar(&c.workers, "workers", 4, "native host workers")
	f.StringVar(&c.backend, "backend", "cuda", "cuda (server target) or cpp (reference)")
	f.StringVar(&c.devices, "devices", "0", "same-node visible CUDA ordinals")
	f.StringVar(&c.library, "gpu-library", "", "trusted CUDA plugin path")
	f.StringVar(&c.layout, "vector-layout", "peer", "peer or local; local adds redistribution traffic")
	f.StringVar(&c.cache, "cache", "", "trusted persistent cache directory")
	f.StringVar(&c.profile, "profile", "", "fresh CPU profile path; default: unique file in cache")
	f.Uint64Var(&c.arrayBytes, "max-array-bytes", 0, "required principal host array cap; not aggregate RSS")
	f.Uint64Var(&c.nativeBytes, "native-bytes", 0, "required native executor cap")
	f.Uint64Var(&c.gpuBytes, "gpu-bytes", 0, "required owned payload cap per GPU")
	f.Uint64Var(&c.codeBytes, "max-code-bytes", 0, "required final canonical payload cap")
	f.Uint64Var(&c.spectrumTile, "spectrum-tile-bytes", 0, "explicit streamed CUDA export tile (8..67108864); zero keeps resident default")
	f.DurationVar(&c.checkpointInterval, "checkpoint-interval", 0, "opt-in recovery save interval (e.g. 2h); sparse writes target 1% overhead")
	f.BoolVar(&c.dimensionOnly, "dimension-only", false, "compute exact K only; no support/grid/preparation arrays")
	f.BoolVar(&c.resume, "resume", false, "reuse completed matching stages in this trusted cache")
	if e := f.Parse(args); e != nil {
		return c, e
	}
	if c.checkpointInterval < 0 || f.NArg() != 0 || c.s < 2 || c.s > 17 || c.t < 1 || c.t >= (1<<uint(c.s))-1 || c.workers < 1 || c.workers > 64 || c.cache == "" || c.arrayBytes == 0 || c.nativeBytes == 0 || (c.codeBytes == 0 && !c.dimensionOnly) || (c.backend != "cuda" && c.backend != "cpp") || (c.layout != "peer" && c.layout != "local") {
		return c, fmt.Errorf("explicit valid specification/backend/cache/byte limits required")
	}
	if c.backend == "cuda" && (c.library == "" || c.gpuBytes == 0) {
		return c, fmt.Errorf("CUDA library and per-device cap required")
	}
	if c.spectrumTile != 0 && (c.spectrumTile < 8 || c.spectrumTile > 64<<20 || c.backend != "cuda" || c.dimensionOnly) {
		return c, fmt.Errorf("spectrum tile requires CUDA export,8bytes..64MiB; not dimension-only")
	}
	return c, nil
}
func devices(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	ids := make([]int, len(parts))
	seen := map[int]bool{}
	if len(parts) > 8 || len(parts)&(len(parts)-1) != 0 {
		return nil, fmt.Errorf("same-node GPU count must be1/2/4/8; cross-node execution unavailable")
	}
	for i, p := range parts {
		v, e := strconv.Atoi(p)
		if e != nil || v < 0 || seen[v] {
			return nil, fmt.Errorf("distinct nonnegative GPU ordinals required")
		}
		seen[v] = true
		ids[i] = v
	}
	return ids, nil
}
func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGUSR1 {
				checkpointRequested.Store(true)
				continue
			}
			pprof.StopCPUProfile()
			_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"utc": time.Now().UTC().Format(time.RFC3339Nano), "phase": "workflow", "event": "terminated", "remaining_work": "unknown"})
			os.Exit(143)
		}
	}()
	if e := run(os.Args[1:], os.Stdout, os.Stderr); e != nil {
		fmt.Fprintln(os.Stderr, e)
		if errors.Is(e, errCheckpointed) {
			os.Exit(75)
		}
		os.Exit(1)
	}
}
func run(args []string, out, events io.Writer) error {
	c, e := parse(args)
	if e != nil {
		return e
	}
	ids, e := devices(c.devices)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(c.cache, 0700); e != nil {
		return e
	}
	lock, e := os.OpenFile(filepath.Join(c.cache, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return e
	}
	defer lock.Close()
	if e = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		return fmt.Errorf("cache already in use: %w", e)
	}
	stages := []string{"support", "root", "code"}
	if !c.resume {
		recovery, _ := filepath.Glob(filepath.Join(c.cache, "recovery-*"))
		if len(recovery) > 0 {
			return fmt.Errorf("recovery stages exist; use --resume")
		}
	}
	if c.dimensionOnly {
		stages = []string{"dimension"}
	}
	for _, stage := range stages {
		if _, err := os.Lstat(filepath.Join(c.cache, stage)); err == nil && !c.resume {
			return fmt.Errorf("cache contains %s; use --resume or a fresh directory", stage)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	var profile *os.File
	if c.profile == "" {
		profile, e = os.CreateTemp(c.cache, "cpu-*.pprof")
	} else {
		profile, e = os.OpenFile(c.profile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	}
	if e != nil {
		return e
	}
	defer profile.Close()
	if e = pprof.StartCPUProfile(profile); e != nil {
		return e
	}
	defer pprof.StopCPUProfile()
	r := &reporter{events, time.Now()}
	r.event("workflow", "started", 0, 1, "operations")
	if e = execute(c, ids, r); e != nil {
		if errors.Is(e, errCheckpointed) {
			r.event("workflow", "checkpointed", 1, 1, "safe-stops")
		} else {
			r.event("workflow", "failed", 0, 1, "operations")
		}
		return e
	}
	r.event("workflow", "completed", 1, 1, "operations")
	kind := "verified preprocessing cache"
	if c.dimensionOnly {
		kind = "exact dimension record (not support/code data)"
	}
	_, e = fmt.Fprintf(out, "%s: %s\nCPU profile: %s\n", kind, c.cache, profile.Name())
	return e
}
func execute(c config, ids []int, r *reporter) error {
	spec := lrs.Spec{FieldDegree: c.s, Dimension: 2, CurveDegree: c.t}
	n := uint64(1) << uint(2*c.s)
	limits := lrs.PreprocessingLimits{MaxElements: n, MaxHostArrayBytes: c.arrayBytes, NativeBytes: c.nativeBytes}
	codeLimits := lrs.ScalingCodeLimits{MaxElements: n, MaxPayloadBytes: c.codeBytes}
	executable, e := os.Executable()
	if e != nil {
		return e
	}

	executableSHA, e := digestPath(executable)
	if e != nil {
		return e
	}
	origin := producer{Revision: sourceRevision, ExecutableSHA256: executableSHA, Backend: c.backend, Devices: c.devices, VectorLayout: c.layout, UTC: time.Now().UTC().Format(time.RFC3339Nano)}
	options := lrs.PreprocessingOptions{Backend: c.backend, Workers: c.workers, Limits: limits, GPUPath: c.library, GPUBytes: c.gpuBytes}
	if c.backend == "cuda" {
		options.GPUDevices = ids
	}
	options.Progress = func(p lrs.PreprocessingProgress) { r.event(p.Phase, p.Event, p.Completed, p.Total, p.Unit) }
	fresh := func(stage string) manifest {
		return manifest{Version: 1, Stage: stage, Spec: spec, Producer: origin, Files: map[string]fileRecord{}}
	}

	if c.dimensionOnly {
		previous, err := readManifest(c.cache, "dimension", spec)
		if err == nil {
			if previous.Dimension == nil || previous.Dimension.N != n || previous.Dimension.K == 0 || previous.Dimension.K >= n {
				return fmt.Errorf("invalid cached exact dimension")
			}
			r.event("exact-dimension-cached", "completed", previous.Dimension.K, n, "monomials")
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		if c.backend == "cuda" {
			origin.GPUPluginSHA256, e = digestPath(c.library)
			if e != nil {
				return e
			}
		}
		if e = os.Setenv("LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT", c.layout); e != nil {
			return e
		}
		k, err := lrs.CountCodeMonomials(spec, options)
		if err != nil {
			return err
		}
		_, err = publish(c.cache, "dimension", fresh("dimension"), func(_ string, m *manifest) error { m.Dimension = &dimensionRecord{K: k, N: n}; return nil })
		if err == nil {
			r.event("exact-dimension-result", "completed", k, n, "monomials")
		}
		return err
	}
	supportManifest, e := readManifest(c.cache, "support", spec)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	rootManifest, rootErr := readManifest(c.cache, "root", spec)
	if rootErr != nil && !os.IsNotExist(rootErr) {
		return rootErr
	}
	codeManifest, codeErr := readManifest(c.cache, "code", spec)
	if codeErr != nil && !os.IsNotExist(codeErr) {
		return codeErr
	}
	if codeErr == nil {
		if e != nil || rootErr != nil {
			return fmt.Errorf("completed code lacks support/root manifests")
		}
		return verifyCompleted(c.cache, spec, supportManifest, rootManifest, codeManifest, codeLimits, r)
	}
	if c.backend == "cuda" {
		origin.GPUPluginSHA256, e = digestPath(c.library)
		if e != nil {
			return e
		}
	}
	// Library selector is process-scoped: set it before constructing the sole
	// engine. No concurrent construction or implicit CPU/host-backed fallback.
	if e = os.Setenv("LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT", c.layout); e != nil {
		return e
	}
	var store *recoveryStore
	if c.checkpointInterval > 0 {
		store = &recoveryStore{cache: c.cache, spec: spec, limits: limits, interval: c.checkpointInterval, last: time.Now(), reporter: r, fresh: fresh, now: time.Now}
		recovery := &lrs.PreprocessingRecovery{Save: store.save}
		if c.resume {
			if supportManifest.Version == 0 {
				recovery.Discovery, e = store.load("discovery")
			} else if rootManifest.Version == 0 {
				recovery.Preparation, e = store.load("roots")
			} else {
				recovery.Preparation, e = store.load("reciprocal")
			}
			if e != nil {
				return e
			}
		}
		options.Recovery = recovery
	}
	var session *lrs.CodePreprocessor
	if e = r.phase("setup", func() error { var err error; session, err = lrs.NewCodePreprocessor(spec, options); return err }); e != nil {
		return e
	}
	defer session.Close()
	var support *lrs.CodeSupport
	if supportManifest.Version == 0 {
		if e = r.phase("discovery", func() error { var err error; support, err = session.Discover(lrs.OrbitOptions{}); return err }); e != nil {
			return e
		}
		supportManifest, e = publish(c.cache, "support", fresh("support"), func(dir string, m *manifest) error {
			width := uint64(4)
			if c.s > 16 {
				width = 8
			}
			var record lrs.SupportRecord
			file, err := writePayload(dir, "support.bin", support.Parameters().K*width, r, func(w io.Writer) (string, error) {
				var err error
				record, err = support.Write(w)
				return record.SHA256, err
			})
			if err != nil {
				return err
			}
			m.Support = &record
			m.Files["support.bin"] = file
			return nil
		})
		if e != nil {
			return e
		}
		if e = verifyFile(c.cache, "support", "support.bin", supportManifest.Files["support.bin"], r); e != nil {
			return e
		}
	} else {
		if supportManifest.Support == nil {
			return fmt.Errorf("missing support record")
		}
		record := supportManifest.Support
		file := supportManifest.Files["support.bin"]
		if file.Bytes != record.Bytes || file.SHA256 != record.SHA256 {
			return fmt.Errorf("support record mismatch")
		}
		if e = withPayload(c.cache, "support", "support.bin", file, r, func(src io.Reader) error {
			var err error
			support, err = lrs.ReadCodeSupport(src, *record, limits)
			return err
		}); e != nil {
			return e
		}
	}
	var root *lrs.RootCheckpoint
	if rootManifest.Version != 0 {
		if rootManifest.Root == nil {
			return fmt.Errorf("missing root record")
		}
		record := rootManifest.Root
		file := rootManifest.Files["root.bin"]
		if file.Bytes != record.Bytes || file.SHA256 != record.SHA256 {
			return fmt.Errorf("root record mismatch")
		}
		if e = withPayload(c.cache, "root", "root.bin", file, r, func(src io.Reader) error {
			var err error
			root, err = lrs.ReadPreparationRoot(src, support, *record, limits)
			return err
		}); e != nil {
			return e
		}
	}
	var prepared *lrs.PreparedCode
	var polynomials *lrs.PreparedPolynomials
	var h lrs.ScalingCodeHeader
	reuse := lrs.PreparationReuse{Root: root}
	if root == nil {
		reuse.SaveRoot = func(root *lrs.RootCheckpoint) error {
			var err error
			rootManifest, err = publish(c.cache, "root", fresh("root"), func(dir string, m *manifest) error {
				var record lrs.RootRecord
				file, err := writePayload(dir, "root.bin", 4*(support.Parameters().K+1), r, func(w io.Writer) (string, error) {
					var err error
					record, err = root.Write(w)
					return record.SHA256, err
				})
				if err != nil {
					return err
				}
				m.Root = &record
				m.Files["root.bin"] = file
				return nil
			})
			if err != nil {
				return err
			}
			if err = verifyFile(c.cache, "root", "root.bin", rootManifest.Files["root.bin"], r); err != nil {
				return err
			}
			if store != nil && checkpointRequested.Load() {
				return errCheckpointed
			}
			return nil
		}
	}
	if e = r.phase("preparation-and-identities", func() error {
		var err error
		if c.spectrumTile == 0 {
			prepared, err = session.Prepare(support, reuse)
			if err == nil {
				h = prepared.CodeData().Header()
			}
		} else {
			polynomials, err = session.PreparePolynomials(support, reuse)
			if err == nil {
				h = polynomials.Header()
			}
		}
		return err
	}); e != nil {
		return e
	}
	if h.Bytes-h.POffset > c.codeBytes {
		return fmt.Errorf("canonical payload exceeds --max-code-bytes")
	}
	if c.spectrumTile != 0 {
		// Only this command's references die. Certified support/root files remain on
		// disk; P/root aliases and G remain owned by the verified polynomial object.
		support = nil
		root = nil
		reuse = lrs.PreparationReuse{}
		if options.Recovery != nil {
			options.Recovery.Discovery = nil
			options.Recovery.Preparation = nil
		}
	}
	r.event("spectrum-tile-bytes", "selected", c.spectrumTile, 64<<20, "bytes; zero=resident")
	// Stream final bytes without a second resident code copy; verify boundedly
	// before atomically publishing the completed stage.
	codeManifest, e = publish(c.cache, "code", fresh("code"), func(dir string, m *manifest) error {
		m.Header = &h
		m.SupportSHA256 = supportManifest.Support.SHA256
		m.RootSHA256 = rootManifest.Root.SHA256
		file, err := writePayload(dir, "online.bin", h.Bytes, r, hashed(func(w io.Writer) error {
			if c.spectrumTile != 0 {
				return session.WriteCode(w, polynomials, int(c.spectrumTile))
			}
			return lrs.WriteScalingCode(w, prepared.CodeData())
		}))
		if err != nil {
			return err
		}
		m.Files["online.bin"] = file
		file, err = writePayload(dir, "generator.bin", 4*(h.N-h.K), r, hashed(func(w io.Writer) error {
			if c.spectrumTile != 0 {
				return polynomials.WriteGenerator(w)
			}
			return prepared.WriteGenerator(w)
		}))
		if err != nil {
			return err
		}
		m.Files["generator.bin"] = file
		// The bounded file verification below no longer needs P/G or a resident word.
		prepared = nil
		polynomials = nil
		if err = verifyFile(dir, ".", "generator.bin", m.Files["generator.bin"], r); err != nil {
			return err
		}
		f, err := os.Open(filepath.Join(dir, "online.bin"))
		if err != nil {
			return err
		}
		defer f.Close()
		return r.phase("verify-final-cache", func() error {
			_, err := lrs.VerifyScalingCode(&byteProgress{reader: f, report: r, phase: "verify-final-cache", total: h.Bytes, last: time.Now()}, m.Files["online.bin"].SHA256, codeLimits)
			return err
		})
	})
	if e != nil {
		return e
	}
	native, device := session.OwnedBytes()
	r.event("native-owned-bytes", "observed", native, c.nativeBytes, "bytes")
	r.event("device-owned-bytes", "observed", device, uint64(len(ids))*c.gpuBytes, "bytes")
	return nil
}
func verifyCompleted(cache string, spec lrs.Spec, support, root, code manifest, limits lrs.ScalingCodeLimits, r *reporter) error {
	if support.Support == nil || root.Root == nil || code.Header == nil {
		return fmt.Errorf("completed cache lacks code/checkpoint records")
	}
	// Validate canonical shape/limits first, before trusting extents from manifests.
	if err := withPayload(cache, "code", "online.bin", code.Files["online.bin"], r, func(src io.Reader) error {
		h, err := lrs.VerifyScalingCode(src, code.Files["online.bin"].SHA256, limits)
		if err != nil {
			return err
		}
		if h != *code.Header {
			return fmt.Errorf("canonical header mismatch")
		}
		return nil
	}); err != nil {
		return err
	}
	if root.Root.SupportSHA256 != support.Support.SHA256 || code.SupportSHA256 != support.Support.SHA256 || code.RootSHA256 != root.Root.SHA256 {
		return fmt.Errorf("completed cache checkpoint binding mismatch")
	}
	sp, rp := support.Support.Parameters, root.Root.Parameters
	sp.Samples = 0
	rp.Samples = 0
	h := code.Header
	if sp != rp || sp.Spec != spec || h.Spec != spec || h.K != sp.K || h.N != sp.N {
		return fmt.Errorf("completed cache specification/count mismatch")
	}
	if sp.Polynomial != h.Polynomial || sp.Delta != h.Delta || sp.Omega != h.Omega {
		return fmt.Errorf("completed cache field mismatch")
	}
	width := uint64(4)
	if spec.FieldDegree > 16 {
		width = 8
	}
	if support.Support.Version != 1 || root.Root.Version != 1 || uint64(support.Support.ElementBytes) != width {
		return fmt.Errorf("completed checkpoint version/width mismatch")
	}
	if support.Support.Bytes != width*h.K || root.Root.Bytes != 4*(h.K+1) || code.Files["generator.bin"].Bytes != 4*(h.N-h.K) || code.Files["online.bin"].Bytes != h.Bytes {
		return fmt.Errorf("completed cache extent mismatch")
	}
	if support.Files["support.bin"] != (fileRecord{support.Support.Bytes, support.Support.SHA256}) || root.Files["root.bin"] != (fileRecord{root.Root.Bytes, root.Root.SHA256}) {
		return fmt.Errorf("checkpoint file record mismatch")
	}
	for _, item := range []struct {
		stage, name string
		m           manifest
	}{{"support", "support.bin", support}, {"root", "root.bin", root}, {"code", "generator.bin", code}} {
		if err := verifyFile(cache, item.stage, item.name, item.m.Files[item.name], r); err != nil {
			return err
		}
	}
	return nil
}
