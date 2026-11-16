package main

import (
	"errors"
	"fmt"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var checkpointRequested atomic.Bool
var errCheckpointed = errors.New("stopped at a saved preprocessing boundary; restart with --resume")
var errCheckpointSlow = errors.New("periodic checkpoint write exceeded its overhead allowance")

// recoveryStore keeps two completed generations per phase. Publication uses the
// same fsync/rename protocol as canonical artifacts. Only failed, unpublished
// recovery temporaries and superseded generations are removed; profiles and
// canonical stages are never pruned. Cache locking is owned by run.
type recoveryStore struct {
	cache    string
	spec     lrs.Spec
	limits   lrs.PreprocessingLimits
	interval time.Duration
	last     time.Time
	spent    time.Duration
	reporter *reporter
	fresh    func(string) manifest
	now      func() time.Time
}

func (s *recoveryStore) generations(kind string) ([]string, error) {
	entries, err := os.ReadDir(s.cache)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "recovery-"+kind+"-") {
			names = append(names, entry.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}
func (s *recoveryStore) load(kind string) (*lrs.PreprocessingCheckpoint, error) {
	names, err := s.generations(kind)
	if err != nil || len(names) == 0 {
		return nil, err
	}
	var lastErr error
	for _, name := range names {
		m, err := readManifest(s.cache, name, s.spec)
		if err == nil && m.Progress == nil {
			err = fmt.Errorf("missing recovery record")
		}
		var checkpoint *lrs.PreprocessingCheckpoint
		if err == nil {
			record := m.Progress
			file := m.Files["progress.bin"]
			if record.Kind != kind || file.Bytes != record.Bytes || file.SHA256 != record.SHA256 {
				err = fmt.Errorf("recovery record/payload mismatch")
			} else {
				err = withPayload(s.cache, name, "progress.bin", file, s.reporter, func(reader io.Reader) error {
					var e error
					checkpoint, e = lrs.ReadPreprocessingCheckpoint(reader, *record, s.limits)
					return e
				})
			}
		}
		if err == nil {
			s.reporter.event("recovery-"+kind, "loaded", 1, 1, "checkpoints")
			return checkpoint, nil
		}
		lastErr = err
		s.reporter.event("recovery-"+kind, "rejected", 0, 1, "checkpoints")
	}
	return nil, fmt.Errorf("no intact %s recovery generation: %w", kind, lastErr)
}

// deadlineWriter checks between bounded writes. It cannot interrupt a blocked
// filesystem syscall or fsync; measured overhead therefore remains a target,
// not a hard latency guarantee. A requested shutdown save bypasses this budget.
type deadlineWriter struct {
	io.Writer
	until time.Time
	now   func() time.Time
}

func (w deadlineWriter) Write(b []byte) (int, error) {
	if w.now().After(w.until) {
		return 0, errCheckpointSlow
	}
	return w.Writer.Write(b)
}

func (s *recoveryStore) save(c *lrs.PreprocessingCheckpoint) error {
	now := s.now()
	forced := checkpointRequested.Load()
	if !forced && now.Sub(s.last) < s.interval {
		return nil
	}
	kind := c.Record().Kind
	stage := "recovery-" + kind + "-" + now.UTC().Format("20060102T150405.000000000")
	// Aim for <=1% cumulative periodic-write overhead. Start after the explicit
	// interval; measured expensive saves stretch the next interval automatically.
	allowance := now.Sub(s.reporter.began)/100 - s.spent
	if allowance < time.Second && !forced {
		return nil
	}
	began := now
	s.reporter.event("recovery-"+kind, "started", 0, 1, "checkpoints")
	_, err := publish(s.cache, stage, s.fresh(stage), func(dir string, m *manifest) error {
		record := c.Record()
		file, err := writePayload(dir, "progress.bin", record.Bytes, s.reporter, func(w io.Writer) (string, error) {
			if !forced {
				w = deadlineWriter{w, began.Add(allowance), s.now}
			}
			var e error
			record, e = c.Write(w)
			return record.SHA256, e
		})
		if err != nil {
			return err
		}
		m.Progress = &record
		m.Files["progress.bin"] = file
		return nil
	})
	cost := s.now().Sub(began)
	s.spent += cost
	s.last = s.now()
	if cost*100 > s.interval {
		s.interval = cost * 100
	}
	if err != nil {
		// Incomplete recovery bytes cannot resume anything. Keeping one per attempt
		// indefinitely could exhaust the explicitly bounded retained-storage budget.
		partial, _ := filepath.Glob(filepath.Join(s.cache, "."+stage+"-*"))
		for _, path := range partial {
			if e := os.RemoveAll(path); e != nil {
				return e
			}
		}
		s.reporter.event("recovery-"+kind, "write-failed", 0, 1, "checkpoints")
		if errors.Is(err, errCheckpointSlow) && !forced {
			return nil
		}
		return err
	}
	names, err := s.generations(kind)
	if err != nil {
		return err
	}
	for _, name := range names[min(2, len(names)):] {
		if err = os.RemoveAll(filepath.Join(s.cache, name)); err != nil {
			return err
		}
	}
	if err = syncDir(s.cache); err != nil {
		return err
	}
	s.reporter.event("recovery-"+kind, "completed", 1, 1, "checkpoints")
	if forced {
		return errCheckpointed
	}
	return nil
}
