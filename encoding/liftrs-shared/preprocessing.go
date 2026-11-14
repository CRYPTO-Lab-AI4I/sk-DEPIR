package liftrs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// PreprocessingLimits admits explicit arrays, not total process RSS. MaxElements
// caps the field grid before allocation. MaxHostArrayBytes caps each principal
// Go array (including grid/support), not their sum or GC retention. NativeBytes
// caps native payload/metadata; allocator/runtime/cache/headroom are additional.
// Full jobs need separately reviewed aggregate host/device/storage budgets.
type PreprocessingLimits struct {
	MaxElements, MaxHostArrayBytes, NativeBytes uint64
}

// PreprocessingOptions keeps server backend choice independent of code parameters.
// Backend is explicit: cpp, or cuda where supported. GPUBytes is mandatory for
// cuda; ordinary CPU client builds do not acquire a CUDA dependency. ProductCutoff
// defaults to4096; smaller products use shared base-field polynomial kernels.
type PreprocessingOptions struct {
	Backend       string
	Workers       int
	Limits        PreprocessingLimits
	GPUPath       string
	GPUBytes      uint64
	ProductCutoff int
	Progress      func(PreprocessingProgress)
}

// PreprocessingProgress measures work inside one named stage, not an overall ETA.
// The callback runs synchronously and must not reenter the preprocessor.
type PreprocessingProgress struct {
	UTC            string  `json:"utc"`
	Phase          string  `json:"phase"`
	Event          string  `json:"event"`
	Completed      uint64  `json:"completed"`
	Total          uint64  `json:"total"`
	Unit           string  `json:"unit"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
}

type preparationReporter struct {
	sink  func(PreprocessingProgress)
	start time.Time
	last  map[string]time.Time
}

func newPreparationReporter(sink func(PreprocessingProgress)) *preparationReporter {
	if sink == nil && os.Getenv("LIFTRS_CPU_PROGRESS") == "1" {
		sink = func(p PreprocessingProgress) { _ = json.NewEncoder(os.Stderr).Encode(p) }
	}
	return &preparationReporter{sink: sink, start: time.Now(), last: make(map[string]time.Time)}
}
func (p *preparationReporter) event(stage, event string, done, total uint64, unit string) {
	if p.sink == nil {
		return
	}
	now := time.Now()
	if event == "progress" && !p.last[stage].IsZero() && now.Sub(p.last[stage]) < 30*time.Second && done != total {
		return
	}
	p.last[stage] = now
	p.sink(PreprocessingProgress{UTC: now.UTC().Format(time.RFC3339Nano), Phase: stage, Event: event, Completed: done, Total: total, Unit: unit, ElapsedSeconds: now.Sub(p.start).Seconds()})
}
func (p *preparationReporter) notice(stage string, done, total uint64, unit string) {
	p.event(stage, "progress", done, total, unit)
}
func (p *preparationReporter) run(stage string, fn func() error) error {
	p.event(stage, "started", 0, 1, "operations")
	if err := fn(); err != nil {
		p.event(stage, "failed", 0, 1, "operations")
		return err
	}
	p.event(stage, "completed", 1, 1, "operations")
	return nil
}

type codePreparationBackend interface {
	GoodMonomials(int, uint64) ([]int, error)
	RootFactorsInto([]int, [][4]uint32) error
	BaseGrid32([]uint32) error
	BaseGrid64([]uint64) error
	Apply32([]uint32, bool) error
	Apply64([]uint64, bool) error
	Product([]uint32, []uint32, int, int) ([]uint32, error)
	Spectrum32([]uint32, []uint32) error
	Spectrum64([]uint32, []uint64) error
	ReleaseProducts()
	NativeOwnedBytes() uint64
	DeviceOwnedBytes() uint64
	Close() error
}

// CodePreprocessor is a single-caller server preparation session. Width is
// selected from s; discovery/root/polynomial logic is shared across backends.
// Close waits for active calls and releases native/device workspace. Produced
// support/code data is Go-owned and remains valid. Do not copy a live session.
// This experimental API has not demonstrated a full s17 preparation run.
type CodePreprocessor struct {
	mu      sync.Mutex
	header  ScalingCodeHeader
	field   wideOrbitField
	options PreprocessingOptions
	backend codePreparationBackend
	closed  bool
}

// NewCodePreprocessor constructs only bounded metadata. CPU native supports
// s2..17, m2; the existing CUDA execution adapter remains s<=16. No automatic
// CPU/streamed fallback is selected for an unavailable requested backend.
func NewCodePreprocessor(spec Spec, options PreprocessingOptions) (*CodePreprocessor, error) {
	s := spec.FieldDegree
	if s < 2 || s > 17 || spec.Dimension != 2 || options.Workers < 1 || options.Workers > 64 ||
		(options.Backend != "cpp" && options.Backend != "cuda") {
		return nil, fmt.Errorf("invalid preprocessing specification/backend/workers")
	}
	n := uint64(1) << uint(2*s)
	h, f, err := canonicalScalingCodeHeader(s, spec.CurveDegree, 1, n)
	if err != nil {
		return nil, err
	}
	width := uint64(h.ExtensionBytes)
	limits := options.Limits
	if limits.MaxElements < n || limits.MaxHostArrayBytes < max(n*width, uint64(1<<20)) || limits.NativeBytes == 0 || n > uint64(^uint(0)>>1)/8 {
		return nil, fmt.Errorf("preprocessing count/host-array/native limits not admitted")
	}
	if options.Backend == "cuda" && options.GPUBytes == 0 {
		return nil, fmt.Errorf("explicit preprocessing GPU byte cap required")
	}
	if options.ProductCutoff == 0 {
		options.ProductCutoff = 4096
	}
	if options.ProductCutoff < 24 {
		return nil, fmt.Errorf("preprocessing product cutoff must be at least24")
	}
	backend, err := makeCodePreparationBackend(h, options)
	if err != nil {
		return nil, err
	}
	return &CodePreprocessor{header: h, field: f, options: options, backend: backend}, nil
}

// SupportParameters describes a certified/trusted reusable code support, not a
// byte-authenticated proof of code membership. Counts/coordinates remain64-bit.
type SupportParameters struct {
	Spec       Spec   `json:"spec"`
	K          uint64 `json:"k"`
	N          uint64 `json:"n"`
	Polynomial uint32 `json:"polynomial"`
	Delta      uint32 `json:"delta"`
	Omega      uint64 `json:"omega"`
	Samples    int    `json:"samples"`
}

// CodeSupport contains immutable sorted exponent indices. A trusted persisted
// manifest must originate from certified discovery; a checksum alone cannot
// prove exact K or support membership. Write/Read helpers preserve source identity.
type CodeSupport struct {
	parameters SupportParameters
	exponents  []int
	checksumMu sync.Mutex
	checksum   string
}

// Parameters returns code identifiers and discovery provenance by value.
func (s *CodeSupport) Parameters() SupportParameters {
	if s == nil {
		return SupportParameters{}
	}
	return s.parameters
}

// Checksum returns the persisted support digest, or empty before publication.
func (s *CodeSupport) Checksum() string {
	if s == nil {
		return ""
	}
	s.checksumMu.Lock()
	defer s.checksumMu.Unlock()
	return s.checksum
}
func (s *CodeSupport) setChecksum(value string) {
	s.checksumMu.Lock()
	s.checksum = value
	s.checksumMu.Unlock()
}

// Discover enumerates the exact good monomial count and samples full codewords
// until the observed univariate support has precisely that size. A bounded
// failure is an error, never an approximate code. Random defaults to crypto/rand.
func (p *CodePreprocessor) Discover(options OrbitOptions) (support *CodeSupport, err error) {
	if p == nil {
		return nil, fmt.Errorf("nil preprocessor")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, fmt.Errorf("preprocessor closed")
	}
	defer recoverPreparation(&err)
	options, err = normalizedOrbitOptions(options)
	if err != nil {
		return nil, err
	}
	reporter := newPreparationReporter(p.options.Progress)
	var good []int
	if err = reporter.run("exact-good-monomials", func() error {
		var e error
		good, e = p.backend.GoodMonomials(p.header.Spec.CurveDegree, p.options.Limits.MaxHostArrayBytes)
		return e
	}); err != nil {
		return nil, err
	}
	var indices []int
	var samples int
	if p.header.ExtensionBytes == 4 {
		evaluate := func(grid []uint32) error {
			if e := reporter.run("discovery-base-grid", func() error { return p.backend.BaseGrid32(grid) }); e != nil {
				return e
			}
			return reporter.run("discovery-extension-interpolation", func() error { return p.backend.Apply32(grid, true) })
		}
		indices, samples, err = discoverSupport(p.header.Spec, good, options, evaluate, reporter.notice)
	} else {
		evaluate := func(grid []uint64) error {
			if e := reporter.run("discovery-base-grid", func() error { return p.backend.BaseGrid64(grid) }); e != nil {
				return e
			}
			return reporter.run("discovery-extension-interpolation", func() error { return p.backend.Apply64(grid, true) })
		}
		indices, samples, err = discoverSupport(p.header.Spec, good, options, evaluate, reporter.notice)
	}
	if err != nil {
		return nil, err
	}
	h := p.header
	return &CodeSupport{parameters: SupportParameters{h.Spec, uint64(len(indices)), h.N, h.Polynomial, h.Delta, h.Omega, samples}, exponents: indices}, nil
}

type preparationConvolver struct{ backend codePreparationBackend }

func (c preparationConvolver) canConvolve(n int) bool { return n >= 4 }
func (c preparationConvolver) convolve(a, b []uint32, limit, n int) []uint32 {
	result, err := c.backend.Product(a, b, limit, n)
	if err != nil {
		panic(cppFailure{err})
	}
	return result
}
func recoverPreparation(err *error) {
	if value := recover(); value != nil {
		if failure, ok := value.(cppFailure); ok {
			*err = failure.err
		} else {
			panic(value)
		}
	}
}

// PreparationReuse requires a trusted root checkpoint bound to this support.
// SaveRoot runs after normalization and before reciprocal work; it must persist
// atomically, must not mutate/retain mutable aliases, and may return an error.
// Reuse always repeats the full P*G identity check before code publication.
type PreparationReuse struct {
	Root     *RootCheckpoint
	SaveRoot func(*RootCheckpoint) error
}

// PreparedCode owns immutable verified P/G/spectrum data independently of the
// preprocessing session. CodeData is directly consumable by CPU/client backends.
// G is retained for reproducibility; WriteGenerator streams it without a copy.
type PreparedCode struct {
	code      *ScalingCodeData
	generator []uint32
	root      *RootCheckpoint
}

// CodeData returns immutable verified data owned independently of the session.
func (p *PreparedCode) CodeData() *ScalingCodeData {
	if p == nil {
		return nil
	}
	return p.code
}

// WriteGenerator streams base-field G coefficients in little-endian uint32 slots.
func (p *PreparedCode) WriteGenerator(destination io.Writer) error {
	if p == nil || p.code == nil {
		return fmt.Errorf("nil prepared code")
	}
	return writeScalingCodeValues(destination, 4, uint64(1)<<uint(p.code.header.Spec.FieldDegree), p.generator, nil)
}

// Prepare constructs normalized P, G=1/P mod X^(N-K), checks the complete
// P*G=1+X^(N-1) identity, then computes the canonical spectrum. It does not
// discover a replacement support, publish files or encode any database.
func (p *CodePreprocessor) Prepare(support *CodeSupport, reuse PreparationReuse) (result *PreparedCode, err error) {
	if p == nil {
		return nil, fmt.Errorf("nil preprocessor")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, fmt.Errorf("preprocessor closed")
	}
	defer recoverPreparation(&err)
	if support == nil {
		return nil, fmt.Errorf("certified support required")
	}
	info, h := support.parameters, p.header
	if info.Spec != h.Spec || info.N != h.N || info.Polynomial != h.Polynomial || info.Delta != h.Delta || info.Omega != h.Omega || info.K != uint64(len(support.exponents)) || info.K > h.N/2 {
		return nil, fmt.Errorf("support differs from preprocessing/code-data shape")
	}
	h, f, err := canonicalScalingCodeHeader(h.Spec.FieldDegree, h.Spec.CurveDegree, info.K, h.N)
	if err != nil {
		return nil, err
	}
	var saved []uint32
	if reuse.Root != nil {
		if err = reuse.Root.matches(support); err != nil {
			return nil, err
		}
		saved = reuse.Root.values
	}
	if reuse.SaveRoot != nil && support.Checksum() == "" {
		return nil, fmt.Errorf("persist support identity before requesting root checkpoint")
	}
	baseField, err := NewBinaryField(info.Spec.FieldDegree, info.Polynomial)
	if err != nil {
		return nil, err
	}
	base := orbitField{q: uint64(1) << uint(info.Spec.FieldDegree), base: baseField, lookup: newPreparationProducts(baseField)}
	base.products = &polynomialProducts{domain: info.N, convolver: preparationConvolver{p.backend}, cutoff: p.options.ProductCutoff}
	reporter := newPreparationReporter(p.options.Progress)
	currentStage := ""
	checkpoint := func(stage string) {
		if currentStage != "" {
			reporter.event(currentStage, "completed", 1, 1, "operations")
		}
		currentStage = strings.TrimPrefix(stage, "before-")
		reporter.event(currentStage, "started", 0, 1, "operations")
	}
	var root *RootCheckpoint
	save := func(values []uint32) error {
		root = &RootCheckpoint{parameters: info, supportSHA: support.Checksum(), values: values}
		if reuse.SaveRoot != nil {
			return reuse.SaveRoot(root)
		}
		return nil
	}
	normalized, g, err := prepareCyclicPolynomials(base, support.exponents, int(info.N), base.q, p.backend.RootFactorsInto, saved, save, checkpoint, reporter.notice)
	if err != nil {
		reporter.event(currentStage, "failed", 0, 1, "operations")
		return nil, err
	}
	reporter.event(currentStage, "completed", 1, 1, "operations")
	if root == nil {
		root = reuse.Root
	}
	// Product operands are dead. Release them before allocating final spectrum;
	// keep only reusable transpose scratch. No recomputation/transfer is added.
	p.backend.ReleaseProducts()
	code := &ScalingCodeData{header: h, field: f, p: normalized}
	if h.ExtensionBytes == 4 {
		code.spectrum32 = make([]uint32, int(h.N))
		err = reporter.run("generator-spectrum", func() error { return p.backend.Spectrum32(g, code.spectrum32) })
	} else {
		code.spectrum64 = make([]uint64, int(h.N))
		err = reporter.run("generator-spectrum", func() error { return p.backend.Spectrum64(g, code.spectrum64) })
	}
	if err != nil {
		return nil, err
	}
	return &PreparedCode{code: code, generator: g, root: root}, nil
}

// OwnedBytes reports explicit native/device payloads. Go arrays and overhead
// remain additional; this is not an aggregate host-memory or scheduler guarantee.
func (p *CodePreprocessor) OwnedBytes() (native, device uint64) {
	if p == nil {
		return 0, 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, 0
	}
	return p.backend.NativeOwnedBytes(), p.backend.DeviceOwnedBytes()
}
func (p *CodePreprocessor) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	return p.backend.Close()
}
