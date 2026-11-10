package cryptoref

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"reflect"
	"sync"
	"testing"
)

func fixedKey(value byte) []byte { return bytes.Repeat([]byte{value}, 32) }

func TestPermutationEntireDomains(t *testing.T) {
	for _, length := range []int{2, 16, 17, 64, 256} {
		p, err := NewPermutation(length, fixedKey(1))
		if err != nil {
			t.Fatal(err)
		}
		positions := make([]int, length)
		for i := range positions {
			positions[i] = i
		}
		mapped, err := p.Map(positions)
		if err != nil {
			t.Fatal(err)
		}
		seen := make([]bool, length)
		for i, position := range mapped {
			if position < 0 || position >= length || seen[position] {
				t.Fatal("not bijective")
			}
			seen[position] = true
			if positions[i] != i {
				t.Fatal("input mutated")
			}
		}
		d, err := NewPermutation(length, fixedKey(1))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p, d) {
			t.Fatal("same key did not reproduce permutation")
		}
		repeated, err := p.Map([]int{0, 0, length - 1})
		if err != nil || repeated[0] != mapped[0] || repeated[1] != mapped[0] || repeated[2] != mapped[length-1] {
			t.Fatal("order or multiplicity lost", err)
		}
		mapped[0] = -1
		again, err := p.Map([]int{0})
		if err != nil || again[0] < 0 {
			t.Fatal("output aliases permutation", err)
		}
	}
}

// Check the prototype helper's index serialization and mask extraction against
// AES directly, then check scatter orientation and XOR independently. AES(0)
// under the 128-bit zero key ends in ...ca342b2e, so its low four bits are 14.
func TestMaskFormatAndXOR(t *testing.T) {
	const length = 16
	key := make([]byte, 16)
	p, err := NewPermutation(length, fixedKey(2))
	if err != nil {
		t.Fatal(err)
	}
	for degree := 2; degree <= 4; degree++ {
		m, err := NewMasker(degree, length, key)
		if err != nil {
			t.Fatal(err)
		}
		q := uint32(1) << uint(degree)
		block, err := aes.NewCipher(key)
		if err != nil {
			t.Fatal(err)
		}
		for position := 0; position < length; position++ {
			var input, output [16]byte
			binary.LittleEndian.PutUint64(input[:8], uint64(position))
			block.Encrypt(output[:], input[:])
			want := uint32(output[15]) & (q - 1)
			got, err := m.Unmask([]int{position}, []uint32{0})
			if err != nil || got[0] != want {
				t.Fatalf("degree %d position %d: mask %v want %d, error %v", degree, position, got, want, err)
			}
		}
		for value := uint32(0); value < q; value++ {
			word := make([]uint32, length)
			for i := range word {
				word[i] = value
			}
			stored, err := m.Encode(word, p)
			if err != nil {
				t.Fatal(err)
			}
			for codePosition, storagePosition := range p.forward {
				var input, output [16]byte
				binary.LittleEndian.PutUint64(input[:8], uint64(storagePosition))
				block.Encrypt(output[:], input[:])
				if stored[storagePosition] != word[codePosition]^(uint32(output[15])&(q-1)) {
					t.Fatal("wrong scatter position or non-XOR masking")
				}
				if word[codePosition] != value {
					t.Fatal("Encode modified input")
				}
			}
			positions, err := p.Map([]int{0, 1, 1, length - 1})
			if err != nil {
				t.Fatal(err)
			}
			values := make([]uint32, len(positions))
			for i, pos := range positions {
				values[i] = stored[pos]
			}
			before := append([]uint32(nil), values...)
			plain, err := m.Unmask(positions, values)
			if err != nil || !reflect.DeepEqual(plain, []uint32{value, value, value, value}) {
				t.Fatal("read roundtrip failed", err)
			}
			if !reflect.DeepEqual(values, before) {
				t.Fatal("Unmask modified responses")
			}
		}
	}
	m, err := NewMasker(4, length, key)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := m.Unmask([]int{0}, []uint32{0})
	if err != nil || golden[0] != 14 {
		t.Fatal("AES zero-block mask vector failed", err)
	}
}

func TestSeparateDatabaseAndReencodingKeys(t *testing.T) {
	p, err := NewPermutation(16, fixedKey(3))
	if err != nil {
		t.Fatal(err)
	}
	word := make([]uint32, 16)
	for i := range word {
		word[i] = uint32(i)
	}
	// Keys A/B represent separate databases; key C represents a new encoding of A.
	var encodings [][]uint32
	for _, key := range [][]byte{fixedKey(0), fixedKey(1), fixedKey(2)} {
		m, err := NewMasker(4, 16, key)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := m.Encode(word, p)
		if err != nil {
			t.Fatal(err)
		}
		encodings = append(encodings, stored)
		positions := make([]int, 16)
		for i := range positions {
			positions[i] = i
		}
		plain, err := m.Unmask(positions, stored)
		if err != nil {
			t.Fatal(err)
		}
		for i, pos := range p.forward {
			if plain[pos] != word[i] {
				t.Fatal("separate-key recovery failed")
			}
		}
	}
	for i := range encodings {
		for j := 0; j < i; j++ {
			// Individual masks may coincide in a tiny field. Compare the whole vectors.
			if reflect.DeepEqual(encodings[i], encodings[j]) {
				t.Fatal("fixed distinct keys produced identical mask vectors")
			}
		}
	}
}

func TestKeyGenerationAndOwnership(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatal("wrong generated key size")
	}
	m, err := NewMasker(4, 16, key)
	if err != nil {
		t.Fatal(err)
	}
	permutationKey := fixedKey(7)
	p, err := NewPermutation(16, permutationKey)
	if err != nil {
		t.Fatal(err)
	}
	word := make([]uint32, 16)
	before, err := m.Encode(word, p)
	if err != nil {
		t.Fatal(err)
	}
	for i := range key {
		key[i] ^= 0xff
		permutationKey[i] ^= 0xff
	}
	after, err := m.Encode(word, p)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("retained caller key slice", err)
	}
	before[0] ^= 1
	if reflect.DeepEqual(before, after) {
		t.Fatal("output buffers alias")
	}
}

func TestInvalidInputs(t *testing.T) {
	for _, length := range []int{-1, 0, 1, 257, int(^uint(0) >> 1)} {
		if _, err := NewPermutation(length, fixedKey(0)); err == nil {
			t.Fatal("accepted length", length)
		}
		if _, err := NewMasker(4, length, fixedKey(0)); err == nil {
			t.Fatal("accepted mask length", length)
		}
	}
	for _, key := range [][]byte{nil, {1}, make([]byte, 15), make([]byte, 33)} {
		if _, err := NewPermutation(16, key); err == nil {
			t.Fatal("accepted bad permutation key")
		}
		if _, err := NewMasker(4, 16, key); err == nil {
			t.Fatal("accepted bad mask key")
		}
	}
	for _, degree := range []int{0, 1, 5, 20} {
		if _, err := NewMasker(degree, 16, fixedKey(0)); err == nil {
			t.Fatal("accepted field degree", degree)
		}
	}
	p, err := NewPermutation(16, fixedKey(0))
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewMasker(4, 16, fixedKey(1))
	if err != nil {
		t.Fatal(err)
	}
	for _, positions := range [][]int{{-1}, {16}, {0, 256}} {
		if _, err := p.Map(positions); err == nil {
			t.Fatal("accepted bad codeword index")
		}
	}
	word := make([]uint32, 16)
	if _, err := m.Encode(word[:15], p); err == nil {
		t.Fatal("accepted wrong codeword length")
	}
	if _, err := m.Encode(word, nil); err == nil {
		t.Fatal("accepted nil permutation")
	}
	different, err := NewPermutation(17, fixedKey(0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Encode(word, different); err == nil {
		t.Fatal("accepted mismatched permutation")
	}
	word[0] = 16
	if _, err := m.Encode(word, p); err == nil {
		t.Fatal("accepted non-field codeword")
	}
	for _, input := range []struct {
		positions []int
		values    []uint32
	}{{[]int{0}, nil}, {[]int{-1}, []uint32{0}}, {[]int{16}, []uint32{0}}, {[]int{0}, []uint32{16}}} {
		if got, err := m.Unmask(input.positions, input.values); err == nil || got != nil {
			t.Fatal("accepted invalid response or returned partial result")
		}
	}
	var zero Masker
	if _, err := zero.Encode(make([]uint32, 16), p); err == nil {
		t.Fatal("accepted zero masker")
	}
	var absent *Masker
	if _, err := absent.Unmask(nil, nil); err == nil {
		t.Fatal("accepted nil masker")
	}
	var zeroPermutation Permutation
	if _, err := zeroPermutation.Map(nil); err == nil {
		t.Fatal("accepted zero permutation")
	}
	var absentPermutation *Permutation
	if _, err := absentPermutation.Map(nil); err == nil {
		t.Fatal("accepted nil permutation")
	}
}

func TestConcurrentReadAndEncoding(t *testing.T) {
	p, err := NewPermutation(64, fixedKey(0))
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewMasker(3, 64, fixedKey(1))
	if err != nil {
		t.Fatal(err)
	}
	word := make([]uint32, 64)
	for i := range word {
		word[i] = uint32(i % 8)
	}
	want, err := m.Encode(word, p)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := m.Encode(word, p)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Error("concurrent encoding failed", err)
				return
			}
			positions, err := p.Map([]int{0, 7, 7})
			if err != nil {
				t.Error(err)
				return
			}
			values := []uint32{got[positions[0]], got[positions[1]], got[positions[2]]}
			plain, err := m.Unmask(positions, values)
			if err != nil || !reflect.DeepEqual(plain, []uint32{0, 7, 7}) {
				t.Error("concurrent unmask failed", err)
			}
		}()
	}
	wg.Wait()
}
