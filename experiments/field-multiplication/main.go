// field-multiplication compares portable kernels with canonical polynomial-basis
// inputs. All variants include modular reduction; no lookup-table setup is hidden.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
	"math/rand"
	"os"
	"runtime"
	"testing"
)

// Original is the preceding production algorithm, retained as a timing baseline.
func original(a, b uint32, s uint, p uint32) uint32 {
	var r uint32
	for b != 0 {
		if b&1 != 0 {
			r ^= a
		}
		b >>= 1
		a <<= 1
		if a&(1<<s) != 0 {
			a ^= p
		}
	}
	return r
}

// Masked uses a fixed number of shift/reduce steps and unsigned all-bit masks.
// This source-level branch removal is not a constant-time security claim.
func masked(a, b uint32, s uint, p uint32) uint32 {
	var r uint32
	for i := uint(0); i < s; i++ {
		r ^= a * (b & 1)
		b >>= 1
		a = (a << 1) ^ (p * (a >> (s - 1)))
	}
	return r
}

// Reduction contains x^s through x^(s+3) reduced modulo p, for each nibble.
func reduction(s uint, p uint32) (r [16]uint32) {
	for i := uint32(0); i < 16; i++ {
		x := i << s
		for j := s + 3; j >= s; j-- {
			x ^= (p << (j - s)) * ((x >> j) & 1)
		}
		r[i] = x
	}
	return
}

// Nibble constructs multiples of a once per multiplication (included in timing),
// then evaluates b in radix x^4. The reduction table is reused per field (64 B).
func nibble(a, b uint32, s uint, p uint32, r *[16]uint32) uint32 {
	var t [16]uint32
	t[1] = a
	t[2] = a << 1
	t[4] = a << 2
	t[8] = a << 3
	for i := 1; i < 16; i++ {
		if i&(i-1) != 0 {
			low := i & -i
			t[i] = t[low] ^ t[i^low]
		}
	}
	var v uint32
	mask := uint32(1<<s) - 1
	for j := int((s+3)/4) - 1; j >= 0; j-- {
		// v is reduced at loop entry; after the XOR its degree is at most s+2.
		x := v<<4 ^ t[(b>>uint(4*j))&15]
		v = x&mask ^ r[x>>s]
	}
	return v
}

var sink uint32

func main() {
	testing.Init()
	flag.Set("test.benchtime", "80ms")
	flag.Parse()
	random := rand.New(rand.NewSource(808))
	var rows []map[string]any
	for s := uint(2); s <= 20; s++ {
		p, _ := liftrs.BinaryPolynomial(int(s))
		field, err := liftrs.NewBinaryField(int(s), p)
		if err != nil {
			panic(err)
		}
		var inputs [4096][2]uint32
		for i := range inputs {
			inputs[i] = [2]uint32{random.Uint32() & ((1 << s) - 1), random.Uint32() & ((1 << s) - 1)}
		}
		table := reduction(s, p)
		variants := []struct {
			name string
			mul  func(uint32, uint32) uint32
		}{
			{"original", func(a, b uint32) uint32 { return original(a, b, s, p) }},
			{"masked", func(a, b uint32) uint32 { return masked(a, b, s, p) }},
			{"nibble", func(a, b uint32) uint32 { return nibble(a, b, s, p, &table) }},
			{"production", field.Mul},
		}
		for _, v := range variants {
			for _, x := range inputs {
				if got, want := v.mul(x[0], x[1]), original(x[0], x[1], s, p); got != want {
					panic(fmt.Sprintf("%s s=%d got=%d want=%d", v.name, s, got, want))
				}
			}
			var samples []float64
			for repeat := 0; repeat < 3; repeat++ {
				b := testing.Benchmark(func(b *testing.B) {
					var result uint32
					for i := 0; i < b.N; i++ {
						x := inputs[i&4095]
						result ^= v.mul(x[0], x[1])
					}
					sink = result
				})
				samples = append(samples, float64(b.T.Nanoseconds())/float64(b.N))
			}
			rows = append(rows, map[string]any{"s": s, "variant": v.name, "ns_per_operation": samples, "operand_setup_included": true})
		}
		setup := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				table = reduction(s, p)
			}
			sink = table[15]
		})
		rows = append(rows, map[string]any{"s": s, "variant": "nibble_reduction_setup", "ns_per_operation": float64(setup.T.Nanoseconds()) / float64(setup.N), "table_bytes": 64})
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"go": runtime.Version(), "goarch": runtime.GOARCH, "seed": 808, "input_pairs_per_field": 4096, "rows": rows})
}
