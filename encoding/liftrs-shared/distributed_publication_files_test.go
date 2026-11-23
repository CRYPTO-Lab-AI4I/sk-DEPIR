package liftrs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// This tests file completion/commit semantics, not the polynomial identity;
// the integration checks identity in runtime and compares native output bytes.
func TestWorkflowPublicationCommit(t *testing.T) {
	directory := t.TempDir()
	p, err := newWorkflowPublication(directory, "s4")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeScalingCodeHeader(p.Code, h); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Code.Write(make([]byte, int(h.Bytes)-int(h.POffset))); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Generator.Write(make([]byte, 4*(h.N-h.K))); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Archive.Write(make([]byte, 8*h.K)); err != nil {
		t.Fatal(err)
	}
	sealed, err := p.SealArchive(8 * h.K)
	if err != nil || sealed != fmt.Sprintf("%x", sha256.Sum256(make([]byte, 8*h.K))) {
		t.Fatal("archive seal", err)
	}
	if _, err = p.Archive.Write([]byte{1}); err == nil {
		t.Fatal("sealed archive mutated")
	}
	if _, err = os.Stat(filepath.Join(directory, "s4-code.bin")); !os.IsNotExist(err) {
		t.Fatal("payload published early")
	}
	if _, err = p.Commit(h, false, 2); err == nil {
		t.Fatal("unverified runtime committed")
	}
	bad := h
	bad.Bytes++
	if _, err = p.Commit(bad, true, 2); err == nil {
		t.Fatal("wrong canonical header admitted")
	}
	record, err := p.Commit(h, true, 2)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "s4-publication.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved workflowPublicationRecord
	if err = json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Files) != 3 || saved.Header != h || saved.SupportFormat != "uint64-exponents-le" {
		t.Fatal("manifest binding")
	}
	for _, f := range record.Files {
		data, e := os.ReadFile(filepath.Join(directory, f.Name))
		if e != nil || uint64(len(data)) != f.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != f.SHA256 {
			t.Fatal("payload hash", f, e)
		}
	}
	alias, e := os.Stat(filepath.Join(directory, "s4-support.bin"))
	if e != nil {
		t.Fatal(e)
	}
	archive, e := os.Stat(filepath.Join(directory, "s4-archive.raw"))
	if e != nil || !os.SameFile(alias, archive) {
		t.Fatal("support payload copied instead of linked", e)
	}
	if _, err = p.Commit(h, true, 2); err == nil {
		t.Fatal("duplicate commit")
	}
	if _, err = newWorkflowPublication(directory, "s4"); err == nil {
		t.Fatal("overwrite admitted")
	}
}
func TestWorkflowPublicationIncompleteRetained(t *testing.T) {
	directory := t.TempDir()
	p, err := newWorkflowPublication(directory, "s4")
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Commit(h, true, 2); err == nil {
		t.Fatal("incomplete file committed")
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(directory, "s4-publication.json")); !os.IsNotExist(err) {
		t.Fatal("incomplete manifest visible")
	}
	if _, err = os.Stat(filepath.Join(directory, "s4-code.bin.partial")); err != nil {
		t.Fatal("partial evidence lost", err)
	}
	if _, err = newWorkflowPublication(directory, "../bad"); err == nil {
		t.Fatal("invalid label")
	}
}

func TestWorkflowPublicationWriteFailure(t *testing.T) {
	directory := t.TempDir()
	p, err := newWorkflowPublication(directory, "s4")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if err = p.files[0].Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Code.Write([]byte("failed")); err == nil || !p.failed {
		t.Fatal("write failure not retained")
	}
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Commit(h, true, 2); err == nil {
		t.Fatal("failed writer committed")
	}
	if _, err = os.Stat(filepath.Join(directory, "s4-publication.json")); !os.IsNotExist(err) {
		t.Fatal("failure marker visible")
	}
}

func TestWorkflowPublicationManifestFailure(t *testing.T) {
	directory := t.TempDir()
	p, err := newWorkflowPublication(directory, "s4")
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeScalingCodeHeader(p.Code, h); err != nil {
		t.Fatal(err)
	}
	for _, part := range []struct {
		writer interface{ Write([]byte) (int, error) }
		bytes  uint64
	}{{p.Code, h.Bytes - h.POffset}, {p.Generator, 4 * (h.N - h.K)}, {p.Archive, 8 * h.K}} {
		if _, err = part.writer.Write(make([]byte, part.bytes)); err != nil {
			t.Fatal(err)
		}
	}
	// Simulate an unavailable manifest destination after payload completion.
	if err = os.Mkdir(filepath.Join(directory, "s4-publication.json.partial"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Commit(h, true, 2); err == nil {
		t.Fatal("failed manifest was accepted")
	}
	if _, err = os.Stat(filepath.Join(directory, "s4-publication.json")); !os.IsNotExist(err) {
		t.Fatal("commit marker visible")
	}
	if _, err = os.Stat(filepath.Join(directory, "s4-code.bin")); err != nil {
		t.Fatal("completed payload lost", err)
	}
}
