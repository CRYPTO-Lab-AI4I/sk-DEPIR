package liftrs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// DistributedPreprocessingConfig is explicit server-job policy, not resource
// authorization. Named-array admission is separate from total RSS/Slurm limits.
// CUDA/MPI remain optional build requirements of this experimental entry only.
type DistributedPreprocessingConfig struct {
	Spec             Spec   `json:"spec"`
	Ranks            int    `json:"ranks"`
	Directory        string `json:"directory"`
	Library          string `json:"library"`
	Device           int    `json:"device"`
	DeviceBytes      uint64 `json:"device_bytes"`
	MaxLocalWords    uint64 `json:"max_local_words"`
	MaxGoodWords     uint64 `json:"max_good_words"`
	MaxRootWords     uint64 `json:"max_root_words"`
	MaxArchiveWords  uint64 `json:"max_archive_words"`
	RSSBytes         uint64 `json:"rss_bytes"`
	HostReserveBytes uint64 `json:"host_reserve_bytes"`
	StorageBytes     uint64 `json:"storage_bytes"`
	FileBytes        uint64 `json:"file_bytes"`
	DeadlineSeconds  int    `json:"deadline_seconds"`
	NewtonInput      string `json:"newton_input"`
	Spectrum         string `json:"spectrum"`
	DiscoveryResult  string `json:"discovery_result"`
	ProofLog         string `json:"proof_log"`
	ResumeDirectory  string `json:"resume_directory,omitempty"`
	SourceSHA        string `json:"source_sha256,omitempty"`
	Known            uint64 `json:"known_prefix,omitempty"`
}

// ReadDistributedPreprocessingConfig rejects oversized, unknown or trailing JSON.
func ReadDistributedPreprocessingConfig(reader io.Reader) (c DistributedPreprocessingConfig, err error) {
	data, e := io.ReadAll(io.LimitReader(reader, 65537))
	if e != nil {
		return c, e
	}
	if len(data) > 65536 {
		return c, fmt.Errorf("configuration exceeds64KiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&c); err != nil {
		return c, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return c, fmt.Errorf("trailing configuration")
	}
	if decoder.InputOffset() > 65536 {
		return c, fmt.Errorf("configuration exceeds64KiB")
	}
	return c, c.Validate()
}

// Validate checks scalar/ownership/resource policy without initializing MPI/CUDA.
// These checks do not establish a fullsize host-memory or runtime estimate.
func (c DistributedPreprocessingConfig) Validate() error {
	s := c.Spec.FieldDegree
	if s < 2 || s > 18 || c.Spec.Dimension != 2 || c.Spec.CurveDegree < 1 || c.Spec.CurveDegree >= (1<<uint(s))-1 || c.Ranks < 2 || c.Ranks > 64 || c.Ranks&(c.Ranks-1) != 0 || !filepath.IsAbs(c.Directory) || !filepath.IsAbs(c.Library) || c.Device < 0 || c.DeviceBytes == 0 || c.RSSBytes == 0 || c.HostReserveBytes == 0 || c.HostReserveBytes >= c.RSSBytes || c.StorageBytes == 0 || c.FileBytes == 0 || c.FileBytes > c.StorageBytes || c.DeadlineSeconds < 1 {
		return fmt.Errorf("explicit valid distributed specification, paths and budgets required")
	}
	n := uint64(1) << uint(2*s)
	words := (n + uint64(c.Ranks) - 1) / uint64(c.Ranks)
	if n < uint64(c.Ranks) || c.MaxLocalWords < words || c.MaxGoodWords == 0 || c.MaxGoodWords > n || c.MaxRootWords == 0 || c.MaxRootWords > n || c.MaxArchiveWords > n || 4*words > c.RSSBytes {
		return fmt.Errorf("distributed host-list/shard admission")
	}
	width := uint64(4)
	if s > 16 {
		width = 8
	}
	if 3*words*width > c.DeviceBytes {
		return fmt.Errorf("three resident GPU arrays exceed owned budget")
	}
	if c.DiscoveryResult == "resident" && c.MaxArchiveWords == 0 {
		return fmt.Errorf("resident discovery requires archive budget")
	}
	if c.NewtonInput != "dense" && c.NewtonInput != "prefix" || c.Spectrum != "resident" && c.Spectrum != "stream" || c.DiscoveryResult != "resident" && c.DiscoveryResult != "stream" || c.ProofLog != "full" && c.ProofLog != "trace" && c.ProofLog != "summary" {
		return fmt.Errorf("explicit supported memory/log modes required")
	}
	if c.ResumeDirectory == "" {
		if c.SourceSHA != "" || c.Known != 0 {
			return fmt.Errorf("recovery requires source directory")
		}
	} else {
		if !filepath.IsAbs(c.ResumeDirectory) || filepath.Clean(c.ResumeDirectory) == filepath.Clean(c.Directory) || c.Known > n {
			return fmt.Errorf("recovery namespace/prefix")
		}
		if _, err := expectedDigest(c.SourceSHA); err != nil {
			return err
		}
	}
	return nil
}

// admitFiles runs once K is known, before root/remaining Newton GPU work. Counts
// include all retained compact checkpoints and both hard-link names logically.
// Reserve bounded JSON metadata; profiles/logs live outside this output namespace.
func (c DistributedPreprocessingConfig) admitFiles(h ScalingCodeHeader) error {
	if h.Spec != c.Spec {
		return fmt.Errorf("job code specification binding")
	}
	total := h.Bytes + 4*(h.N-h.K) + 16*h.K
	if max(h.Bytes, 4*(h.N-h.K), 8*h.K) > c.FileBytes {
		return fmt.Errorf("canonical output file budget")
	}
	total += 4 * (h.K + 1)
	_, words := rankBand(h.N, 0, c.Ranks)
	if 4*min(h.K+1, words) > c.FileBytes {
		return fmt.Errorf("root checkpoint file budget")
	}
	stages := uint64(1)
	keep := uint64(1)
	for keep < h.N-h.K {
		keep = min(2*keep, h.N-h.K)
		if c.Known != 0 && keep < c.Known {
			continue
		}
		total += 4 * keep
		stages++
		if 4*min(keep, words) > c.FileBytes {
			return fmt.Errorf("Newton checkpoint file budget")
		}
	}
	// Each source/commit is at most64KiB. Arithmetic payloads are limited to s18.
	total += (stages + 2) * 65536
	if total > c.StorageBytes {
		return fmt.Errorf("retained output/checkpoint budget: need %d bytes, cap %d", total, c.StorageBytes)
	}
	return nil
}
