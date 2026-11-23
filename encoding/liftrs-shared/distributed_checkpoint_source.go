package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type workflowCheckpointSource struct {
	Version                   int               `json:"version"`
	Header                    ScalingCodeHeader `json:"header"`
	Ranks, Samples            int
	SupportFormat, SupportSHA string
	SupportBytes              uint64
}

// Write metadata last through an exclusive temporary file; consumers must ignore
// payloads without a matching commit. The caller exclusively owns this namespace.
// Failures leave evidence in place;filesystem fsync semantics bound durability.
func writeWorkflowMetadata(directory, name string, data []byte) error {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || len(data) > 65536 {
		return fmt.Errorf("checkpoint metadata admission")
	}
	final := filepath.Join(directory, name)
	if _, e := os.Lstat(final); !os.IsNotExist(e) {
		return fmt.Errorf("checkpoint metadata destination exists")
	}
	f, err := os.OpenFile(final+".partial", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if n, e := f.Write(data); e != nil {
		err = e
	} else if n != len(data) {
		err = fmt.Errorf("checkpoint metadata short write")
	} else {
		err = f.Sync()
	}
	ce := f.Close()
	if err != nil {
		return err
	}
	if ce != nil {
		return ce
	}
	if err = syncWorkflowDirectory(directory); err != nil {
		return err
	}
	if err = os.Rename(final+".partial", final); err != nil {
		return err
	}
	return syncWorkflowDirectory(directory)
}

// This checksum binds code coordinates,curve degree,sample count and raw archive
// format,not merely K/N. It is not authentication of untrusted source metadata.
func readWorkflowCheckpointSource(directory, label, digest string) (source workflowCheckpointSource, err error) {
	data, err := readBoundedPilotFile(filepath.Join(directory, label+"-checkpoint-source.json"), 65536)
	if err != nil {
		return source, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		return source, fmt.Errorf("checkpoint source checksum")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&source); err != nil {
		return source, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return source, fmt.Errorf("checkpoint source trailing metadata")
	}

	h := source.Header
	canonical, _, e := canonicalScalingCodeHeader(h.Spec.FieldDegree, h.Spec.CurveDegree, h.K, h.N)
	if e != nil || canonical != h || source.Version != 1 || source.Ranks < 2 || source.Ranks > 64 || source.Ranks&(source.Ranks-1) != 0 || source.Samples < 1 || source.Samples > discoverySampleLimit(h.Spec.FieldDegree) || source.SupportFormat != "uint64-exponents-le" || source.SupportBytes != 8*h.K {
		return source, fmt.Errorf("checkpoint source metadata binding")
	}
	if _, err = expectedDigest(source.SupportSHA); err != nil {
		return source, err
	}
	return source, nil
}
