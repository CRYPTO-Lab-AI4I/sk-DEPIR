package main

import (
	"bytes"
	"errors"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"os"
	"path/filepath"
	"testing"
)

func TestPublicationFailureAndReuse(t *testing.T) {
	cache := t.TempDir()
	spec := lrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 3}
	m := manifest{Version: 1, Stage: "support", Spec: spec}
	injected := errors.New("interrupted before manifest")
	_, err := publish(cache, "support", m, func(dir string, _ *manifest) error {
		if e := os.WriteFile(filepath.Join(dir, "support.bin"), []byte("partial"), 0600); e != nil {
			return e
		}
		return injected
	})
	if !errors.Is(err, injected) {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(cache, "support")); !os.IsNotExist(err) {
		t.Fatal("incomplete stage published")
	}
	entries, err := os.ReadDir(cache)
	if err != nil || len(entries) != 1 {
		t.Fatal("partial evidence not retained", err)
	}
	_, err = publish(cache, "support", m, func(dir string, _ *manifest) error {
		return os.WriteFile(filepath.Join(dir, "support.bin"), []byte("complete"), 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readManifest(cache, "support", spec); err != nil {
		t.Fatal(err)
	}
	if _, err = readManifest(cache, "support", lrs.Spec{FieldDegree: 5, Dimension: 2, CurveDegree: 3}); err == nil {
		t.Fatal("mismatched specification accepted")
	}
	if _, err = publish(cache, "support", m, func(string, *manifest) error { return nil }); err == nil {
		t.Fatal("completed stage replaced")
	}
	raw, _ := os.ReadFile(filepath.Join(cache, "support", "support.bin"))
	if !bytes.Equal(raw, []byte("complete")) {
		t.Fatal("completed bytes altered")
	}
}
func TestExplicitCLIAndDeviceBounds(t *testing.T) {
	for _, value := range []string{"0,0", "-1", "0,1,2", "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15"} {
		if _, err := devices(value); err == nil {
			t.Fatal("invalid device list", value)
		}
	}
	if ids, err := devices("0,1,2,3,4,5,6,7"); err != nil || len(ids) != 8 {
		t.Fatal(ids, err)
	}
	if _, err := parse([]string{"--s", "17"}); err == nil {
		t.Fatal("missing resource/cache flags accepted")
	}
}
