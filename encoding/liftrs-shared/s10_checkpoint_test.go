package liftrs

import (
	"math/rand"
	"os"
	"testing"
)

// This moderate-instance fixture supplies a different message from the timing
// experiment. Curve evaluation uses the independent carryless field oracle.
func TestS10StoredArtifact(t *testing.T) {
	data, err := os.ReadFile("../../experiments/encoder-scaling/artifacts/s10-m2-t6-cyclic-fft.json")
	if err != nil {
		t.Fatal(err)
	}
	encoder, err := LoadCyclicEncoder(data)
	if err != nil {
		t.Fatal(err)
	}
	basis := encoder.Basis()
	if basis.Dimension() != 103556 || basis.Length() != 1048576 || basis.polynomial != 0x409 {
		t.Fatal("unexpected s=10 specification")
	}
	count := 0
	for exponent := 0; exponent < 1048576; exponent++ {
		good, err := GoodMonomial(Spec{10, 2, 6}, []uint32{uint32(exponent % 1024), uint32(exponent / 1024)})
		if err != nil {
			t.Fatal(err)
		}
		if good {
			count++
		}
	}
	if count != basis.Dimension() {
		t.Fatal("exact enumeration disagrees", count)
	}
	random := rand.New(rand.NewSource(7311))
	message := make([]uint32, count)
	for i := range message {
		message[i] = uint32(random.Intn(1024))
	}
	word, err := encoder.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	for j, x := range message {
		pos, err := basis.SystematicPosition(j)
		if err != nil || word[pos] != x {
			t.Fatal("systematic mismatch", j, err)
		}
	}
	for trial := 0; trial < 16; trial++ {
		var curve [2][7]uint32
		for axis := range curve {
			for j := range curve[axis] {
				curve[axis][j] = uint32(random.Intn(1024))
			}
		}
		var sum uint32
		for z := uint32(0); z < 1024; z++ {
			var point [2]uint32
			for axis := range curve {
				for j := 6; j >= 0; j-- {
					point[axis] = binaryOracleMul(point[axis], z, 10, 0x409) ^ curve[axis][j]
				}
			}
			sum ^= word[point[0]+1024*point[1]]
		}
		if sum != 0 {
			t.Fatal("sampled degree-six curve", trial)
		}
	}
}
