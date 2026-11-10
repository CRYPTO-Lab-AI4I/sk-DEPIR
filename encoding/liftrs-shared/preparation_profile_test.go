//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"syscall"
	"testing"
	"time"
)

type profileTransform struct {
	delegate *scalingPreparationCPP
	calls    int
	nanos    int64
}

func (p *profileTransform) apply(data []uint32, inverse bool) {
	start := time.Now()
	p.delegate.apply(data, inverse)
	p.calls++
	p.nanos += time.Since(start).Nanoseconds()
}

// This opt-in profile samples public support only. Its bounded executor does
// not allocate the full s16 scratch arena or attempt a partial whole-code P.
func TestPreparationTargetProfile(t *testing.T) {
	out := os.Getenv("PIR_TARGET_OUT")
	if out == "" {
		t.Skip("opt-in targeted preparation profile")
	}
	debug.SetMemoryLimit(4 << 30)
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(signals)
	go func() { <-signals; pprof.StopCPUProfile(); os.Exit(4) }()
	var stages []map[string]any
	phase := func(name string, fn func() map[string]any) {
		runtime.GC()
		f, err := os.Create(filepath.Join(out, name+".pprof"))
		check(err)
		check(pprof.StartCPUProfile(f))
		start := time.Now()
		data := fn()
		elapsed := time.Since(start)
		pprof.StopCPUProfile()
		check(f.Close())
		data["phase"], data["seconds"] = name, elapsed.Seconds()
		stages = append(stages, data)
		raw, err := json.MarshalIndent(stages, "", "  ")
		check(err)
		check(os.WriteFile(filepath.Join(out, "stages.json"), raw, 0644))
		t.Log(name, elapsed)
	}
	var support []int
	phase("support-read", func() map[string]any {
		cache := os.Getenv("PIR_TARGET_SUPPORT")
		raw, err := os.ReadFile(filepath.Join(cache, "basis-metadata.json"))
		check(err)
		var meta struct {
			S, T, K, N int
			SHA256     string
		}
		check(json.Unmarshal(raw, &meta))
		if meta.S != 16 || meta.T != 5 || meta.N != 1<<32 || meta.K != 918672421 || meta.SHA256 != os.Getenv("PIR_TARGET_SHA256") {
			t.Fatal("support provenance mismatch")
		}
		f, err := os.Open(filepath.Join(cache, "support.bin"))
		check(err)
		defer f.Close()
		h := sha256.New()
		r := io.TeeReader(f, h)
		const samples = 65536
		var block [65536]byte
		for start := 0; start < meta.K; start += len(block) / 4 {
			n := minInt(len(block)/4, meta.K-start)
			_, err := io.ReadFull(r, block[:4*n])
			check(err)
			for len(support) < samples {
				index := int(uint64(len(support)) * uint64(meta.K) / samples)
				if index >= start+n {
					break
				}
				v := binary.LittleEndian.Uint32(block[4*(index-start):])
				support = append(support, int(v))
			}
		}
		var extra [1]byte
		n, err := r.Read(extra[:])
		if n != 0 || err != io.EOF {
			t.Fatal("support trailing bytes")
		}
		if hex.EncodeToString(h.Sum(nil)) != meta.SHA256 {
			t.Fatal("support checksum")
		}
		return map[string]any{"sampled": len(support), "source_k": meta.K, "sha256": meta.SHA256, "sampling": "floor(j*K/65536), spread over sorted support"}
	})
	var base, extension orbitField
	var omega uint32
	var fft *additiveTransform
	var meter *profileTransform
	c := &scalingPreparationCPP{plans: make(map[int]*matrixCPPPlan), workers: 4}
	defer c.close()
	phase("bounded-setup", func() map[string]any {
		base, extension, _, omega = orbitFields(16, 0x1002d)
		base.lookup = newPreparationProducts(base.base)
		fft = newSparseTransform(extension)
		for dimension := 2; dimension <= 22; dimension++ {
			p, err := newMatrixCPP(extension, dimension)
			check(err)
			c.plans[1<<uint(dimension)] = p
		}
		c.scratch = make([]uint32, 1<<22)
		meter = &profileTransform{delegate: c}
		fft.executor = meter
		base.products = &polynomialProducts{transform: fft, cutoff: 32768, workspace: &productWorkspace{}}
		return map[string]any{"s": 16, "workers": 4, "maximum_transform_points": 1 << 22, "transpose_bytes": len(c.scratch) * 4}
	})
	var factors [][]uint32
	var roots []uint32
	phase("root-generation", func() map[string]any {
		order := int(extension.q) - 1
		for _, i := range support {
			partner := int(uint64(i) * base.q % uint64(order))
			if partner < i {
				continue
			}
			root := extension.inv(extension.pow(omega, i))
			roots = append(roots, root)
			if partner == i {
				if uint64(root) >= base.q {
					t.Fatal("fixed root")
				}
				factors = append(factors, []uint32{root, 1})
			} else {
				conjugate := extension.pow(root, int(base.q))
				trace, norm := root^conjugate, extension.mul(root, conjugate)
				if uint64(trace) >= base.q || uint64(norm) >= base.q {
					t.Fatal("paired coefficients")
				}
				factors = append(factors, []uint32{norm, trace, 1})
			}
		}
		return map[string]any{"positions_scanned": len(support), "factors": len(factors)}
	})
	var products [][]uint32
	phase("leaf-products", func() map[string]any {
		before := meter.calls
		for start := 0; start < len(factors); start += 8192 {
			products = append(products, rootProduct(base, factors[start:minInt(start+8192, len(factors))]))
		}
		if meter.calls != before {
			t.Fatal("leaf profile unexpectedly used native transforms")
		}
		return map[string]any{"factors": len(factors), "batch_factors": 8192, "batches": len(products)}
	})
	var product []uint32
	phase("batch-merges", func() map[string]any {
		beforeCalls, beforeNS := meter.calls, meter.nanos
		product = rootProduct(base, products)
		return map[string]any{"coefficients": len(product), "native_transform_calls": meter.calls - beforeCalls, "native_transform_seconds": float64(meter.nanos-beforeNS) / 1e9}
	})
	eval := func(coeff []uint32, x uint32) uint32 {
		v := uint32(0)
		for i := len(coeff) - 1; i >= 0; i-- {
			v = extension.mul(v, x) ^ coeff[i]
		}
		return v
	}
	phase("factor-verification", func() map[string]any {
		for i, factor := range factors {
			if eval(factor, roots[i]) != 0 {
				t.Fatal("factor root mismatch", i)
			}
		}
		for _, x := range []uint32{0, 1, omega} {
			want := uint32(1)
			for _, factor := range factors {
				want = extension.mul(want, eval(factor, x))
			}
			if eval(product, x) != want {
				t.Fatal("factor product evaluation")
			}
		}
		return map[string]any{"roots_checked": len(roots), "product_evaluation_points": 3}
	})
	for _, logN := range []uint{18, 20, 22} {
		n := 1 << logN
		a, b := make([]uint32, n/2), make([]uint32, n/2)
		for i := range a {
			a[i] = uint32((i*173 + 7) & 65535)
			b[i] = uint32((i*337 + 13) & 65535)
		}
		var result []uint32
		phase(fmt.Sprintf("convolution-%d", logN), func() map[string]any {
			beforeCalls, beforeNS := meter.calls, meter.nanos
			result = polyMul(base, a, b)
			return map[string]any{"transform_points": n, "output_coefficients": len(result), "native_transform_calls": meter.calls - beforeCalls, "native_transform_seconds": float64(meter.nanos-beforeNS) / 1e9}
		})
		phase(fmt.Sprintf("verify-convolution-%d", logN), func() map[string]any {
			for _, v := range result {
				if uint64(v) >= base.q {
					t.Fatal("convolution left base field")
				}
			}
			for _, x := range []uint32{0, 1, omega} {
				if eval(result, x) != extension.mul(eval(a, x), eval(b, x)) {
					t.Fatal("convolution evaluation")
				}
			}
			// Independent direct convolution of the low coefficients.
			for i := 0; i < 32; i++ {
				want := uint32(0)
				for j := 0; j <= i; j++ {
					want ^= base.base.Mul(a[j], b[i-j])
				}
				if result[i] != want {
					t.Fatal("convolution prefix")
				}
			}
			return map[string]any{"evaluation_points": 3, "direct_prefix_checked": 32, "base_range_checked": len(result)}
		})
	}
}
