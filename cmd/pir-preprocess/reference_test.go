//go:build liftrs_cpp && cgo

package main

import (
	"bytes"
	"encoding/json"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise actual publication/restart with an independent Go arithmetic oracle;
// changing random samples cannot change the certified support or canonical code.
func TestPersistentReferenceAndRestart(t *testing.T) {
	cache := t.TempDir()
	args := []string{"--s", "4", "--t", "3", "--backend", "cpp", "--workers", "2", "--cache", cache, "--max-array-bytes", "1048576", "--native-bytes", "8388608", "--max-code-bytes", "4194304"}
	var events bytes.Buffer
	if err := run(args, io.Discard, &events); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(cache, "code", "online.bin"))
	if err != nil {
		t.Fatal(err)
	}
	basis, err := lrs.DiscoverOrbit(lrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 3}, lrs.OrbitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	encoder, err := basis.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := encoder.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	if err = lrs.WriteScalingOnline(&expected, plan); err != nil || !bytes.Equal(first, expected.Bytes()) {
		t.Fatal("independent Go canonical bytes", err)
	}
	if err = run(args, io.Discard, io.Discard); err == nil {
		t.Fatal("unrequested overwrite accepted")
	}
	args = append(args, "--resume")
	events.Reset()
	if err = run(args, io.Discard, &events); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(events.String(), `"phase":"setup"`) {
		t.Fatal("completed-cache resume created compute engine")
	}

	rootPath := filepath.Join(cache, "root", "manifest.json")
	originalRoot, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	var badRoot manifest
	if err = json.Unmarshal(originalRoot, &badRoot); err != nil {
		t.Fatal(err)
	}
	badRoot.Root.Parameters.Delta ^= 1
	altered, err := json.Marshal(badRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(rootPath, altered, 0600); err != nil {
		t.Fatal(err)
	}
	if err = run(args, io.Discard, io.Discard); err == nil {
		t.Fatal("mismatched root parameters accepted")
	}
	if err = os.WriteFile(rootPath, originalRoot, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(filepath.Join(cache, "code"), filepath.Join(cache, "code-first")); err != nil {
		t.Fatal(err)
	}
	events.Reset()
	if err = run(args, io.Discard, &events); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(events.String(), `"phase":"discovery"`) {
		t.Fatal("root resume repeated discovery")
	}
	replay, _ := os.ReadFile(filepath.Join(cache, "code", "online.bin"))
	if !bytes.Equal(first, replay) {
		t.Fatal("root checkpoint replay differs")
	}
	var record manifest
	raw, _ := os.ReadFile(filepath.Join(cache, "code", "manifest.json"))
	if err = json.Unmarshal(raw, &record); err != nil || record.SupportSHA256 == "" || record.RootSHA256 == "" {
		t.Fatal("source binding", err)
	}
	replay[len(replay)-1] ^= 1
	if err = os.WriteFile(filepath.Join(cache, "code", "online.bin"), replay, 0600); err != nil {
		t.Fatal(err)
	}
	if err = run(args, io.Discard, io.Discard); err == nil {
		t.Fatal("corrupted cache accepted")
	}
	after, _ := os.ReadFile(filepath.Join(cache, "code", "online.bin"))
	if !bytes.Equal(after, replay) {
		t.Fatal("corrupt evidence altered")
	}
}
