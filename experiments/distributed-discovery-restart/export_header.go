//go:build ignore

// Validate canonical coordinates without loading preprocessing payload arrays.
package main

import (
	"encoding/json"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"os"
)

func main() {
	f, e := os.Open(os.Args[1])
	if e != nil {
		panic(e)
	}
	defer f.Close()
	h, e := lrs.ReadScalingCodeHeader(f)
	if e != nil {
		panic(e)
	}
	if e = json.NewEncoder(os.Stdout).Encode(h); e != nil {
		panic(e)
	}
}
