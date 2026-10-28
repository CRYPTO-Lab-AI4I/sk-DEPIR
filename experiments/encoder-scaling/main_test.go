package main

import (
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
	"math/rand"
	"testing"
)

func TestCurveOracleArithmetic(t *testing.T) {
	f, err := liftrs.NewBinaryField(10, 0x409)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(7320))
	for i := 0; i < 10000; i++ {
		a, b := uint32(random.Intn(1024)), uint32(random.Intn(1024))
		if oracleMul(a, b) != f.Mul(a, b) {
			t.Fatal("oracle disagreement", a, b)
		}
	}
}
