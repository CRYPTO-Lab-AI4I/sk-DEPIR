//go:build !darwin && liftrs_cpp && cgo

package main

import "os"

// No platform-specific bypass is guessed. Report cached I/O explicitly.
func noCache(f *os.File) (string, error) { return "OS-cached; no bypass implemented", nil }
