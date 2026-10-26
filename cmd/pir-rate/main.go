// pir-rate estimates Lift-RS code rates by uniformly sampling reduced monomials.
// It does not preprocess a code, encode a database, or estimate security.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

type estimate struct {
	ID             string  `json:"id"`
	S              int     `json:"s"`
	M              int     `json:"m"`
	T              int     `json:"t"`
	N              uint64  `json:"n"`
	Seed           int64   `json:"seed"`
	Samples        uint64  `json:"samples"`
	Good           uint64  `json:"good"`
	Rate           float64 `json:"rate"`
	LowerRate      float64 `json:"lower_rate"`
	UpperRate      float64 `json:"upper_rate"`
	Confidence     float64 `json:"confidence"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
}

type results struct {
	Version          int        `json:"version"`
	CatalogueSHA256  string     `json:"catalogue_sha256"`
	GoVersion        string     `json:"go_version"`
	GOOS             string     `json:"goos"`
	GOARCH           string     `json:"goarch"`
	RootSeed         int64      `json:"root_seed"`
	FamilyConfidence float64    `json:"family_confidence"`
	Method           string     `json:"method"`
	Estimates        []estimate `json:"estimates"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "pir-rate:", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("pir-rate", flag.ContinueOnError)
	path := flags.String("catalogue", "pirconfig/configs_catalogue.json", "version-1 PIR configuration catalogue")
	samples := flags.Uint64("samples", 1000000, "uniform monomial samples per configuration")
	seed := flags.Int64("seed", 20261026, "math/rand root seed; reproducible, not cryptographic")
	confidence := flags.Float64("confidence", .99, "simultaneous confidence across all configurations (Hoeffding/union bound)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *samples < 1 || *samples > 1<<53 || math.IsNaN(*confidence) || *confidence <= 0 || *confidence >= 1 {
		return fmt.Errorf("require no positional arguments, 1<=samples<=2^53, and 0<confidence<1")
	}
	raw, err := os.ReadFile(*path)
	if err != nil {
		return err
	}
	c, err := pirconfig.Load(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	hash := sha256.Sum256(raw)
	result := results{Version: 1, CatalogueSHA256: hex.EncodeToString(hash[:]), GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, RootSeed: *seed, FamilyConfidence: *confidence, Method: "Uniform exponent tuples sampled with replacement using Go math/rand.NewSource; exact bounded-carry goodness test. Two-sided Hoeffding intervals with union bound across configurations; confidence assumes independent uniform sampling."}
	// P(|p_hat-p|>epsilon) <= 2 exp(-2*n*epsilon^2). Allocate the
	// family error equally across configurations; intervals need not be independent.
	alpha := (1 - *confidence) / float64(len(c.Configurations))
	margin := math.Sqrt(math.Log(2/alpha) / (2 * float64(*samples)))
	for _, entry := range c.Configurations {
		spec := liftrs.Spec{FieldDegree: entry.Code.S, Dimension: entry.Code.M, CurveDegree: entry.Code.T}
		// Seed by code parameters so reordering entries does not change their samples.
		configSeed := *seed + int64(spec.FieldDegree)*1000 + int64(spec.Dimension)*100 + int64(spec.CurveDegree)
		rng := rand.New(rand.NewSource(configSeed))
		exponents := make([]uint32, spec.Dimension)
		mask := uint32(1)<<uint(spec.FieldDegree) - 1
		var good uint64
		start := time.Now()
		for i := uint64(0); i < *samples; i++ {
			for j := range exponents {
				exponents[j] = rng.Uint32() & mask
			}
			accepted, err := liftrs.GoodMonomial(spec, exponents)
			if err != nil {
				return err
			}
			if accepted {
				good++
			}
		}
		elapsed := time.Since(start).Seconds()
		rate := float64(good) / float64(*samples)
		result.Estimates = append(result.Estimates, estimate{ID: entry.ID, S: spec.FieldDegree, M: spec.Dimension, T: spec.CurveDegree, N: entry.Capacity.N, Seed: configSeed, Samples: *samples, Good: good, Rate: rate, LowerRate: math.Max(0, rate-margin), UpperRate: math.Min(1, rate+margin), Confidence: 1 - alpha, ElapsedSeconds: elapsed})
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
