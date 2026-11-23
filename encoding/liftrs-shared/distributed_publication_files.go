package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

type workflowPublishedFile struct {
	Name   string `json:"name"`
	Bytes  uint64 `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type workflowPublicationRecord struct {
	Version          int                     `json:"version"`
	Header           ScalingCodeHeader       `json:"header"`
	IdentityVerified bool                    `json:"identity_verified"`
	Samples          int                     `json:"samples"`
	SupportFormat    string                  `json:"support_format"`
	Files            []workflowPublishedFile `json:"files"`
}

// The caller exclusively owns this directory/label. Consumers accept files only
// through the last-written commit manifest, never just because payload names exist.
// This detects incomplete output/corruption, not malicious manifest replacement.
type workflowPublication struct {
	directory, label         string
	Code, Generator, Archive io.Writer
	files                    [3]*os.File
	hashes                   [3]hash.Hash
	written                  [3]uint64
	failed                   bool
	closed, committed        bool
}

func newWorkflowPublication(directory, label string) (p *workflowPublication, err error) {
	if label == "" || label == "." || filepath.Base(label) != label {
		return nil, fmt.Errorf("publication label")
	}
	// An exclusive reservation prevents accidental retry/overwrite. Incomplete
	// data stays for investigation; no old output or reservation is deleted.
	reservation, err := os.OpenFile(filepath.Join(directory, label+"-publication.pending"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	if err = reservation.Close(); err != nil {
		return nil, err
	}
	p = &workflowPublication{directory: directory, label: label}
	opened := p
	defer func() {
		if err != nil {
			_ = opened.Close()
		}
	}()
	for i, name := range p.names() {
		if _, e := os.Lstat(filepath.Join(directory, name)); !os.IsNotExist(e) {
			return nil, fmt.Errorf("publication destination already exists: %s", name)
		}
		f, e := os.OpenFile(filepath.Join(directory, name+".partial"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if e != nil {
			return nil, e
		}
		p.files[i] = f
		p.hashes[i] = sha256.New()
		sink := workflowPublicationWriter{p, i}
		switch i {
		case 0:
			p.Code = sink
		case 1:
			p.Generator = sink
		case 2:
			p.Archive = sink
		}
	}
	if _, e := os.Lstat(filepath.Join(directory, label+"-publication.json")); !os.IsNotExist(e) {
		return nil, fmt.Errorf("publication manifest already exists")
	}
	return p, nil
}
func (p *workflowPublication) names() []string {
	return []string{p.label + "-code.bin", p.label + "-generator.bin", p.label + "-archive.raw"}
}
func (p *workflowPublication) Close() error {
	if p == nil || p.closed {
		return nil
	}
	p.closed = true
	var first error
	for _, f := range p.files {
		if f != nil {
			if err := f.Close(); err != nil && first == nil {
				first = err
			}
		}
	}
	return first
}
func syncWorkflowDirectory(directory string) error {
	f, err := os.Open(directory)
	if err != nil {
		return err
	}
	err = f.Sync()
	ce := f.Close()
	if err != nil {
		return err
	}
	return ce
}

// Commit syncs exact-size staged payloads, validates their canonical code header,
// uses hashes accumulated during writes, then renames payloads and commits
// the manifest last. IdentityVerified is a result from the runtime, not an
// identity recomputation here. Callers retain trusted source/runtime provenance.
// On any error no success may be advertised; leave all files for diagnosis.
func (p *workflowPublication) Commit(h ScalingCodeHeader, identity bool, samples int) (record workflowPublicationRecord, err error) {
	if p == nil || p.closed || p.committed || p.failed || !identity || samples < 1 || h.K == 0 || h.K >= h.N || h.N > 1<<40 {
		return record, fmt.Errorf("publication completion admission")
	}
	actualHeader, e := ReadScalingCodeHeader(io.NewSectionReader(p.files[0], 0, 65536))
	if e != nil {
		return record, e
	}
	if actualHeader != h {
		return record, fmt.Errorf("publication canonical header binding")
	}
	var header bytes.Buffer
	if err = writeScalingCodeHeader(&header, h); err != nil {
		return record, err
	}
	files := p.files
	sizes := []uint64{h.Bytes, 4 * (h.N - h.K), 8 * h.K}
	names := p.names()
	for i, f := range files {
		if err = f.Sync(); err != nil {
			return record, err
		}
		info, e := f.Stat()
		if e != nil {
			return record, e
		}
		if info.Size() < 0 || uint64(info.Size()) != sizes[i] || p.written[i] != sizes[i] {
			return record, fmt.Errorf("publication file extent: %s", names[i])
		}
		if i == 0 {
			actual := make([]byte, header.Len())
			if _, err = f.ReadAt(actual, 0); err != nil {
				return record, err
			}
			if !bytes.Equal(actual, header.Bytes()) {
				return record, fmt.Errorf("publication code header mismatch")
			}
		}
		record.Files = append(record.Files, workflowPublishedFile{names[i], sizes[i], fmt.Sprintf("%x", p.hashes[i].Sum(nil))})
	}
	if err = p.Close(); err != nil {
		return record, err
	}
	for _, name := range names {
		if err = os.Rename(filepath.Join(p.directory, name+".partial"), filepath.Join(p.directory, name)); err != nil {
			return record, err
		}
	}
	if err = syncWorkflowDirectory(p.directory); err != nil {
		return record, err
	}
	record.Version = 1
	record.Header = h
	record.IdentityVerified = true
	record.Samples = samples
	record.SupportFormat = "uint64-exponents-le"
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return record, err
	}
	temporary := filepath.Join(p.directory, p.label+"-publication.json.partial")
	f, err := os.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return record, err
	}
	if n, e := f.Write(data); e != nil {
		err = e
	} else if n != len(data) {
		err = io.ErrShortWrite
	} else {
		err = f.Sync()
	}
	ce := f.Close()
	if err != nil {
		return record, err
	}
	if ce != nil {
		return record, ce
	}
	if err = os.Rename(temporary, filepath.Join(p.directory, p.label+"-publication.json")); err != nil {
		return record, err
	}
	if err = syncWorkflowDirectory(p.directory); err != nil {
		return record, err
	}
	p.committed = true
	return record, nil
}

// The payload files are private;all normal writes update their checksum/count.
// As with existing streamed caches,this trusts immutable caller-owned storage.
// Later readers must verify the committed hashes against bytes read from disk.
type workflowPublicationWriter struct {
	owner *workflowPublication
	index int
}

func (w workflowPublicationWriter) Write(data []byte) (int, error) {
	p := w.owner
	if p.closed || p.failed {
		return 0, io.ErrClosedPipe
	}
	n, err := p.files[w.index].Write(data)
	if n > 0 {
		p.hashes[w.index].Write(data[:n])
		p.written[w.index] += uint64(n)
	}
	if err != nil {
		p.failed = true
		return n, err
	}
	if n != len(data) {
		p.failed = true
		return n, io.ErrShortWrite
	}
	return n, nil
}
