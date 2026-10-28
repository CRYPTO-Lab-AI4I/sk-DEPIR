package main

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
)

func TestIndependentCurveArithmetic(t *testing.T) {
	random := rand.New(rand.NewSource(9001))
	for s := 2; s <= 20; s++ {
		polynomial, err := liftrs.BinaryPolynomial(s)
		check(err)
		field, err := liftrs.NewBinaryField(s, polynomial)
		check(err)
		for i := 0; i < 1000; i++ {
			a, b := uint32(random.Intn(1<<uint(s))), uint32(random.Intn(1<<uint(s)))
			if oracleMul(a, b, s, polynomial) != field.Mul(a, b) {
				t.Fatal("oracle multiplication", s, a, b)
			}
		}
	}
}

func TestStreamingPacking(t *testing.T) {
	random := rand.New(rand.NewSource(9002))
	for s := 2; s <= 13; s++ {
		for _, n := range []int{0, 1, 7, 19, 100} {
			word := make([]uint32, n)
			packed := make([]byte, (n*s+7)/8)
			for i := range word {
				word[i] = uint32(random.Intn(1 << uint(s)))
				for bit := 0; bit < s; bit++ {
					if word[i]&(1<<uint(bit)) != 0 {
						position := i*s + bit
						packed[position/8] |= 1 << uint(position%8)
					}
				}
			}
			hash, size, _ := packedOutput(word, s)
			want := sha256.Sum256(packed)
			if hash != hex.EncodeToString(want[:]) || size != int64(len(packed)) {
				t.Fatal("streaming packing", s, n)
			}
		}
	}
}
