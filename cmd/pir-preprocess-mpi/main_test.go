package main

import (
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandResources(t *testing.T) {
	c := lrs.DistributedPreprocessingConfig{Directory: t.TempDir(), DeadlineSeconds: 30, RSSBytes: 1 << 40, FileBytes: 16, StorageBytes: 20}
	if err := resources(c, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c.Directory, "a"), make([]byte, 17), 0600); err != nil {
		t.Fatal(err)
	}
	if err := resources(c, time.Now()); err == nil {
		t.Fatal("file guard")
	}
	c.FileBytes = 20
	if err := os.WriteFile(filepath.Join(c.Directory, "b"), make([]byte, 4), 0600); err != nil {
		t.Fatal(err)
	}
	if err := resources(c, time.Now()); err == nil {
		t.Fatal("storage guard")
	}
	c.StorageBytes = 100
	c.RSSBytes = 1
	if err := resources(c, time.Now()); err == nil {
		t.Fatal("RSS guard")
	}
	c.RSSBytes = 1 << 40
	if err := resources(c, time.Now().Add(-time.Minute)); err == nil {
		t.Fatal("deadline")
	}
}
