package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"path/filepath"
	"time"
)

type fileRecord struct {
	Bytes  uint64 `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type producer struct {
	GPUPluginSHA256  string `json:"gpu_plugin_sha256,omitempty"`
	Revision         string `json:"revision"`
	ExecutableSHA256 string `json:"executable_sha256"`
	Backend          string `json:"backend"`
	Devices          string `json:"devices"`
	VectorLayout     string `json:"vector_layout"`
	UTC              string `json:"utc"`
}

// Manifests certify only artifacts produced in this trusted local cache. A hash
// supplied alongside arbitrary data is not a mathematical certificate.
type dimensionRecord struct {
	K uint64 `json:"k"`
	N uint64 `json:"n"`
}

type manifest struct {
	Dimension     *dimensionRecord       `json:"dimension,omitempty"`
	Version       int                    `json:"version"`
	Stage         string                 `json:"stage"`
	Spec          lrs.Spec               `json:"spec"`
	Producer      producer               `json:"producer"`
	Files         map[string]fileRecord  `json:"files"`
	Support       *lrs.SupportRecord     `json:"support,omitempty"`
	Root          *lrs.RootRecord        `json:"root,omitempty"`
	Header        *lrs.ScalingCodeHeader `json:"header,omitempty"`
	SupportSHA256 string                 `json:"support_sha256,omitempty"`
	RootSHA256    string                 `json:"root_sha256,omitempty"`
}

type reporter struct {
	out   io.Writer
	began time.Time
}

func (r *reporter) event(phase, event string, done, total uint64, unit string) {
	_ = json.NewEncoder(r.out).Encode(map[string]any{"utc": time.Now().UTC().Format(time.RFC3339Nano), "phase": phase, "event": event, "completed": done, "total": total, "unit": unit, "elapsed_seconds": time.Since(r.began).Seconds()})
}
func (r *reporter) phase(name string, fn func() error) error {
	r.event(name, "started", 0, 1, "operations")
	if err := fn(); err != nil {
		r.event(name, "failed", 0, 1, "operations")
		return err
	}
	r.event(name, "completed", 1, 1, "operations")
	return nil
}

type byteProgress struct {
	reader      io.Reader
	writer      io.Writer
	report      *reporter
	phase       string
	done, total uint64
	last        time.Time
}

func (p *byteProgress) advance(n int) {
	p.done += uint64(n)
	if time.Since(p.last) >= 30*time.Second {
		p.report.event(p.phase, "progress", p.done, p.total, "bytes")
		p.last = time.Now()
	}
}
func (p *byteProgress) Read(b []byte) (int, error) {
	n, e := p.reader.Read(b)
	p.advance(n)
	return n, e
}
func (p *byteProgress) Write(b []byte) (int, error) {
	n, e := p.writer.Write(b)
	p.advance(n)
	return n, e
}

func syncDir(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	return f.Sync()
}
func readManifest(cache, stage string, spec lrs.Spec) (manifest, error) {
	var m manifest
	f, e := os.Open(filepath.Join(cache, stage, "manifest.json"))
	if e != nil {
		return m, e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return m, e
	}
	if !stat.Mode().IsRegular() || stat.Size() > 1<<20 {
		return m, fmt.Errorf("manifest size/type limit")
	}
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	if e = d.Decode(&m); e != nil {
		return m, e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return m, fmt.Errorf("manifest trailing data")
	}
	if m.Version != 1 || m.Stage != stage || m.Spec != spec {
		return m, fmt.Errorf("%s manifest specification/version mismatch", stage)
	}
	return m, nil
}

// A directory rename publishes payloads and manifest together. Failed partial
// directories are retained for diagnosis, never mistaken for completed stages.
func publish(cache, stage string, m manifest, write func(string, *manifest) error) (manifest, error) {
	if _, e := os.Lstat(filepath.Join(cache, stage)); e == nil {
		return m, fmt.Errorf("stage already exists: %s", stage)
	} else if !os.IsNotExist(e) {
		return m, e
	}
	dir, e := os.MkdirTemp(cache, "."+stage+"-")
	if e != nil {
		return m, e
	}
	if e = write(dir, &m); e != nil {
		return m, e
	}
	raw, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return m, e
	}
	f, e := os.OpenFile(filepath.Join(dir, "manifest.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return m, e
	}
	_, e = f.Write(append(raw, '\n'))
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return m, e
	}
	if e = syncDir(dir); e != nil {
		return m, e
	}
	if e = os.Rename(dir, filepath.Join(cache, stage)); e != nil {
		return m, e
	}
	return m, syncDir(cache)
}

// The producer supplies the digest if its codec already hashes; avoid a second
// checksum pass over support/root. Other payloads hash as they are streamed.
func writePayload(dir, name string, total uint64, r *reporter, produce func(io.Writer) (string, error)) (fileRecord, error) {
	var record fileRecord
	e := r.phase("write-"+name, func() error {
		f, e := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		p := &byteProgress{writer: f, report: r, phase: "write-" + name, total: total, last: time.Now()}
		digest, writeErr := produce(p)
		if writeErr == nil {
			writeErr = f.Sync()
		}
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		if p.done != total {
			return fmt.Errorf("%s wrote %d bytes; expected %d", name, p.done, total)
		}
		record = fileRecord{p.done, digest}
		r.event("write-"+name, "progress", p.done, total, "bytes")
		return nil
	})
	return record, e
}
func hashed(write func(io.Writer) error) func(io.Writer) (string, error) {
	return func(w io.Writer) (string, error) {
		h := sha256.New()
		e := write(io.MultiWriter(w, h))
		return hex.EncodeToString(h.Sum(nil)), e
	}
}
func withPayload(cache, stage, name string, record fileRecord, r *reporter, read func(io.Reader) error) error {
	if record.Bytes == 0 || len(record.SHA256) != 64 {
		return fmt.Errorf("missing %s payload identity", name)
	}
	return r.phase("read-"+stage+"-"+name, func() error {
		f, e := os.Open(filepath.Join(cache, stage, name))
		if e != nil {
			return e
		}
		defer f.Close()
		stat, e := f.Stat()
		if e != nil {
			return e
		}
		if !stat.Mode().IsRegular() || uint64(stat.Size()) != record.Bytes {
			return fmt.Errorf("%s size/type mismatch", name)
		}
		p := &byteProgress{reader: f, report: r, phase: "read-" + stage + "-" + name, total: record.Bytes, last: time.Now()}
		if e = read(p); e != nil {
			return e
		}
		if p.done != record.Bytes {
			return fmt.Errorf("incomplete %s read", name)
		}
		r.event(p.phase, "progress", p.done, p.total, "bytes")
		return nil
	})
}
func verifyFile(cache, stage, name string, record fileRecord, r *reporter) error {
	return withPayload(cache, stage, name, record, r, func(src io.Reader) error {
		h := sha256.New()
		if _, e := io.Copy(h, src); e != nil {
			return e
		}
		if hex.EncodeToString(h.Sum(nil)) != record.SHA256 {
			return fmt.Errorf("%s checksum mismatch", name)
		}
		return nil
	})
}

func digestPath(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
