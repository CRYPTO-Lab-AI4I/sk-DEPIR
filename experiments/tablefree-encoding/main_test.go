package main

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestPackingMatchesBitDefinition(t *testing.T) {
	r := rand.New(rand.NewSource(32))
	for s := 2; s <= 10; s++ {
		for n := 0; n < 40; n++ {
			word := make([]uint32, n)
			want := make([]byte, (n*s+7)/8)
			for i := range word {
				word[i] = uint32(r.Intn(1 << uint(s)))
				for bit := 0; bit < s; bit++ {
					if word[i]&(1<<uint(bit)) != 0 {
						index := i*s + bit
						want[index/8] |= 1 << uint(index%8)
					}
				}
			}
			if !reflect.DeepEqual(packWord(word, s), want) {
				t.Fatal("packed output differs", s, n)
			}
		}
	}
}
