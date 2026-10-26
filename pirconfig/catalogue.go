// Package pirconfig loads versioned Lift-RS configuration catalogues without
// running code preprocessing or encoding. Validation checks internal consistency;
// evidence and backend support still require review before executing a plan.
package pirconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"reflect"
	"regexp"
)

// Catalogue is the version-1 document for the full-field curve-lifted construction.
// Partition plans reference configuration IDs; they are separate documents.
type Catalogue struct {
	Version        int             `json:"version"`
	Construction   string          `json:"construction"`
	Configurations []Configuration `json:"configurations"`
}

// Configuration records a code, query profile, evidence, and declared encoding
// availability. Candidate configurations may be useful before capacity is known.
type Configuration struct {
	ID             string        `json:"id"`
	EncodingStatus string        `json:"encoding_status"` // candidate or supported
	Code           Code          `json:"code"`
	Capacity       Capacity      `json:"capacity"`
	Query          Query         `json:"query"`
	Security       Security      `json:"security"`
	Encoder        *Encoder      `json:"encoder,omitempty"`
	Measurements   []Measurement `json:"measurements,omitempty"`
}

// Code specifies Lift-RS_q(m,q-1,t). Polynomial is the coefficient bitmask
// including X^s; omission means the field representation has not been selected.
type Code struct {
	S          int     `json:"s"`
	M          int     `json:"m"`
	T          int     `json:"t"`
	Polynomial *uint32 `json:"polynomial,omitempty"`
}

// Capacity separates unknown, estimated, and verified message dimensions.
// Estimated K is a rounded point estimate, not permission to encode K records.
// Rate is K/N; application padding and measured storage overhead are separate.
type Capacity struct {
	Status   string     `json:"status"`
	N        uint64     `json:"n"`
	K        *uint64    `json:"k,omitempty"`
	Rate     *float64   `json:"rate,omitempty"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

// Evidence links repository-root-relative paths (or durable URLs) to the source
// of a claim. Monte Carlo evidence includes the method and uncertainty interval.
// Links are recorded without opening files, executing scripts, or fetching URLs.
type Evidence struct {
	Kind       string      `json:"kind"`
	Reference  string      `json:"reference"`
	Script     string      `json:"script,omitempty"`
	Results    string      `json:"results,omitempty"`
	Method     string      `json:"method,omitempty"`
	MonteCarlo *MonteCarlo `json:"monte_carlo,omitempty"`
}

// MonteCarlo describes a rate confidence interval, including sampling size.
// Confidence is a fraction in (0,1); the method must explain its interpretation.
type MonteCarlo struct {
	Samples    uint64  `json:"samples"`
	Confidence float64 `json:"confidence"`
	LowerRate  float64 `json:"lower_rate"`
	UpperRate  float64 `json:"upper_rate"`
}

// Query counts every transmitted position, including repetitions and dummies.
// Version 1 uses q-1 curve positions and a fixed number of random dummy positions.
type Query struct {
	CurvePositions uint64 `json:"curve_positions"`
	DummyPositions uint64 `json:"dummy_positions"`
	Length         uint64 `json:"length"`
	DummyRule      string `json:"dummy_rule"`
}

// Security records a documented estimate, not a security certification. A value
// of zero bits is used for correctness fixtures with no security claim.
type Security struct {
	ComputationalBits uint32     `json:"computational_bits"`
	Log2Samples       uint32     `json:"log2_samples"`
	Assumptions       string     `json:"assumptions"`
	Evidence          []Evidence `json:"evidence"`
}

// Encoder identifies the implementation and reusable code-preprocessing method.
// Loading does not invoke either method or register an executable backend.
type Encoder struct {
	Implementation string     `json:"implementation"`
	Preprocessing  string     `json:"preprocessing"`
	Evidence       []Evidence `json:"evidence"`
}

// Measurement reports an existing resource measurement with reproducible evidence.
// Setup is charged once per distinct code; encoding is per PIR database. These
// observations are informational and must not influence partition selection.
type Measurement struct {
	Stage    string   `json:"stage"`    // code_preprocessing or database_encoding
	Quantity string   `json:"quantity"` // elapsed_seconds, peak_bytes, artifact_bytes, stored_bytes
	Value    float64  `json:"value"`
	Evidence Evidence `json:"evidence"`
}

var identifier = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Load decodes one JSON document, rejecting duplicate keys, unknown fields,
// explicit nulls, trailing documents, and inconsistent configurations. The caller
// owns the returned mutable data and may impose an input-size limit on r.
func Load(r io.Reader) (*Catalogue, error) {
	if r == nil {
		return nil, fmt.Errorf("pirconfig: nil reader")
	}
	// Inspect keys before decoding: encoding/json otherwise accepts duplicates and
	// null scalar fields, silently replacing declarations with zero values.
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pirconfig: read: %w", err)
	}
	if err := checkJSON(raw); err != nil {
		return nil, fmt.Errorf("pirconfig: %w", err)
	}
	if err := checkShape(raw, reflect.TypeOf(Catalogue{})); err != nil {
		return nil, fmt.Errorf("pirconfig: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var c Catalogue
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("pirconfig: decode: %w", err)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// Validate checks arithmetic, evidence metadata, and declared support. It does
// not prove capacity, authenticate evidence, or guarantee that a referenced
// backend exists in the caller's build. Encoding must also check backend support.
func (c *Catalogue) Validate() error {
	if c == nil || c.Version != 1 || c.Construction != "curve-lifted-rs-qminus1" {
		return fmt.Errorf("pirconfig: require version 1 and curve-lifted-rs-qminus1 construction")
	}
	if len(c.Configurations) == 0 {
		return fmt.Errorf("pirconfig: configurations must be nonempty")
	}
	seen := make(map[string]bool)
	for i := range c.Configurations {
		entry := &c.Configurations[i]
		if !identifier.MatchString(entry.ID) || seen[entry.ID] {
			return fmt.Errorf("pirconfig: invalid or duplicate id %q", entry.ID)
		}
		seen[entry.ID] = true
		if err := entry.validate(); err != nil {
			return fmt.Errorf("pirconfig: %s: %w", entry.ID, err)
		}
	}
	return nil
}

func (c *Configuration) validate() error {
	spec := c.Code
	// These bounds cover the target s=8..20,m=2 and the existing tiny fixtures.
	// s*m <= 60 keeps dimensions within uint64 and JSON's integer interchange
	// caveat is documented for consumers using binary64 numbers.
	if spec.S < 2 || spec.S > 20 || spec.M < 2 || spec.M > 3 || spec.T < 1 || spec.T >= (1<<uint(spec.S))-1 {
		return fmt.Errorf("unsupported code specification")
	}
	n := uint64(1) << uint(spec.S*spec.M)
	if c.Capacity.N != n {
		return fmt.Errorf("n must equal 2^(s*m)")
	}
	if spec.Polynomial != nil && !irreducible(*spec.Polynomial, spec.S) {
		return fmt.Errorf("polynomial must be irreducible of degree s")
	}
	if c.EncodingStatus != "candidate" && c.EncodingStatus != "supported" {
		return fmt.Errorf("invalid encoding_status")
	}
	cap := c.Capacity
	switch cap.Status {
	case "unknown":
		if cap.K != nil || cap.Rate != nil {
			return fmt.Errorf("unknown capacity must omit k and rate")
		}
	case "estimated", "verified":
		if cap.K == nil || *cap.K == 0 || *cap.K > n || cap.Rate == nil || !finite(*cap.Rate) || *cap.Rate <= 0 || *cap.Rate > 1 || math.Abs(*cap.Rate-float64(*cap.K)/float64(n)) > 1e-12 {
			return fmt.Errorf("capacity requires 0 < k <= n and rate = k/n")
		}
		if len(cap.Evidence) == 0 {
			return fmt.Errorf("known capacity requires evidence")
		}
	default:
		return fmt.Errorf("invalid capacity status")
	}
	verified := false
	for _, e := range cap.Evidence {
		if err := e.validate(); err != nil {
			return err
		}
		if e.Kind == "encoding" || e.Kind == "reference_test" {
			verified = true
		}
		if e.MonteCarlo != nil && cap.Rate != nil && (*cap.Rate < e.MonteCarlo.LowerRate-1e-12 || *cap.Rate > e.MonteCarlo.UpperRate+1e-12) {
			return fmt.Errorf("rate lies outside Monte Carlo interval")
		}
	}
	if cap.Status == "verified" && !verified {
		return fmt.Errorf("verified capacity requires encoding or reference_test evidence")
	}
	q := uint64(1) << uint(spec.S)
	if c.Query.CurvePositions != q-1 || c.Query.DummyRule != "fixed_count_uniform_positions" || c.Query.DummyPositions > ^uint64(0)-(q-1) || c.Query.Length != q-1+c.Query.DummyPositions {
		return fmt.Errorf("query must contain q-1 curve positions plus fixed-count dummies")
	}
	if c.Security.Assumptions == "" || len(c.Security.Evidence) == 0 {
		return fmt.Errorf("security requires assumptions and evidence")
	}
	for _, e := range c.Security.Evidence {
		if err := e.validate(); err != nil {
			return err
		}
	}
	if c.Encoder != nil {
		if c.Encoder.Implementation == "" || c.Encoder.Preprocessing == "" || len(c.Encoder.Evidence) == 0 {
			return fmt.Errorf("encoder requires implementation, preprocessing, and evidence")
		}
		for _, e := range c.Encoder.Evidence {
			if err := e.validate(); err != nil {
				return err
			}
		}
	}
	if c.EncodingStatus == "supported" && (cap.Status != "verified" || spec.Polynomial == nil || c.Encoder == nil) {
		return fmt.Errorf("supported encoding requires verified capacity, a polynomial, and an encoder")
	}
	for _, m := range c.Measurements {
		if m.Stage != "code_preprocessing" && m.Stage != "database_encoding" {
			return fmt.Errorf("invalid measurement stage")
		}
		switch m.Quantity {
		case "elapsed_seconds", "peak_bytes", "artifact_bytes", "stored_bytes":
		default:
			return fmt.Errorf("invalid measurement quantity")
		}
		if !finite(m.Value) || m.Value < 0 || m.Evidence.Kind != "measurement" {
			return fmt.Errorf("measurement requires nonnegative finite value and measurement evidence")
		}
		if err := m.Evidence.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e Evidence) validate() error {
	switch e.Kind {
	case "paper", "analytical", "monte_carlo", "encoding", "reference_test", "measurement":
	default:
		return fmt.Errorf("invalid evidence kind %q", e.Kind)
	}
	if e.Reference == "" {
		return fmt.Errorf("evidence requires a reference")
	}
	if e.Kind == "analytical" || e.Kind == "monte_carlo" || e.Kind == "encoding" || e.Kind == "measurement" {
		if e.Script == "" || e.Results == "" {
			return fmt.Errorf("%s evidence requires script and results links", e.Kind)
		}
	}
	if e.Kind == "monte_carlo" {
		m := e.MonteCarlo
		if m == nil || e.Method == "" || m.Samples == 0 || !finite(m.Confidence) || m.Confidence <= 0 || m.Confidence >= 1 || !finite(m.LowerRate) || !finite(m.UpperRate) || m.LowerRate < 0 || m.UpperRate > 1 || m.LowerRate > m.UpperRate {
			return fmt.Errorf("Monte Carlo evidence requires method, samples, confidence, and a valid rate interval")
		}
	} else if e.MonteCarlo != nil {
		return fmt.Errorf("monte_carlo metadata requires monte_carlo kind")
	}
	return nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Rabin's criterion: gcd(X^(2^i)-X,P)=1 for i<=s/2 and X^(2^s)=X mod P.
// Polynomial coefficients are bits; subtraction is XOR.
func irreducible(p uint32, s int) bool {
	if p>>uint(s) != 1 || p&1 == 0 {
		return false
	}
	x := uint32(2)
	for i := 1; i <= s; i++ {
		var square uint64
		for j := 0; j < s; j++ {
			if x&(1<<uint(j)) != 0 {
				square ^= 1 << uint(2*j)
			}
		}
		x = uint32(polyRem(square, uint64(p)))
		if i <= s/2 {
			a, b := uint64(p), uint64(x^2)
			for b != 0 {
				a, b = b, polyRem(a, b)
			}
			if a != 1 {
				return false
			}
		}
	}
	return x == 2
}

func polyRem(a, b uint64) uint64 {
	for a != 0 && bits.Len64(a) >= bits.Len64(b) {
		a ^= b << uint(bits.Len64(a)-bits.Len64(b))
	}
	return a
}
