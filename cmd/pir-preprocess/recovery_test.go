//go:build liftrs_cpp && cgo

package main

import (
	"bytes"
	"errors"
	"fmt"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Trigger a signal-equivalent request from a real stage event, then start a fresh
// CLI session on the persisted state. This exercises atomic publication, lookup,
// parent bindings, native work and canonical artifact publication together.
type requestOnEvent struct {
	kind   string
	events bytes.Buffer
}

func (w *requestOnEvent) Write(p []byte) (int, error) {
	w.events.Write(p)
	if strings.Contains(string(p), `"phase":"`+w.kind+`"`) {
		checkpointRequested.Store(true)
	}
	return len(p), nil
}
func TestCLIRecoveryRestart(t *testing.T) {
	defer checkpointRequested.Store(false)
	for _, phase := range []string{"discovery-support-union", "root-product", "reciprocal-prefix"} {
		t.Run(phase, func(t *testing.T) {
			checkpointRequested.Store(false)
			cache := t.TempDir()
			args := []string{"--s", "4", "--t", "3", "--backend", "cpp", "--workers", "2", "--cache", cache, "--max-array-bytes", "1048576", "--native-bytes", "8388608", "--max-code-bytes", "4194304", "--checkpoint-interval", "2h"}
			events := &requestOnEvent{kind: phase}
			if e := run(args, io.Discard, events); !errors.Is(e, errCheckpointed) {
				t.Fatal("expected safe stop", e, events.events.String())
			}
			checkpointRequested.Store(false)
			args = append(args, "--resume")
			var replay bytes.Buffer
			if e := run(args, io.Discard, &replay); e != nil {
				t.Fatal("resume", e, replay.String())
			}
			if !strings.Contains(replay.String(), `"event":"loaded"`) {
				t.Fatal("did not load recovery")
			}
			if _, e := os.Stat(filepath.Join(cache, "code", "online.bin")); e != nil {
				t.Fatal(e)
			}
			// Running again verifies bytes with no executor, as before recovery support.
			if e := run(args, io.Discard, io.Discard); e != nil {
				t.Fatal("completed readback", e)
			}
		})
	}
}

func TestRecoveryDeadlineWriter(t *testing.T) {
	now := time.Now()
	var payload bytes.Buffer
	w := deadlineWriter{Writer: &payload, until: now, now: func() time.Time { return now.Add(time.Second) }}
	if _, err := w.Write([]byte("partial")); !errors.Is(err, errCheckpointSlow) || payload.Len() != 0 {
		t.Fatal("overhead guard", err)
	}
}

func TestRecoveryRollingRetentionAndFallback(t *testing.T) {
	defer checkpointRequested.Store(false)
	spec := lrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 3}
	limits := lrs.PreprocessingLimits{MaxElements: 256, MaxHostArrayBytes: 1 << 20, NativeBytes: 8 << 20}
	cache := t.TempDir()
	now := time.Now()
	report := &reporter{out: io.Discard, began: now.Add(-3 * time.Hour)}
	store := &recoveryStore{cache: cache, spec: spec, limits: limits, interval: 2 * time.Hour, last: now.Add(-3 * time.Hour), reporter: report, now: func() time.Time { return now }, fresh: func(stage string) manifest {
		return manifest{Version: 1, Stage: stage, Spec: spec, Files: map[string]fileRecord{}}
	}}
	recovery := &lrs.PreprocessingRecovery{}
	session, err := lrs.NewCodePreprocessor(spec, lrs.PreprocessingOptions{Backend: "cpp", Workers: 2, Limits: limits, Recovery: recovery})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	stops := fmt.Errorf("test stopped")
	recovery.Save = func(c *lrs.PreprocessingCheckpoint) error {
		for i := 0; i < 3; i++ {
			checkpointRequested.Store(true)
			if err := store.save(c); !errors.Is(err, errCheckpointed) {
				return err
			}
			now = now.Add(time.Second)
		}
		return stops
	}
	if _, err = session.Discover(lrs.OrbitOptions{}); !errors.Is(err, stops) {
		t.Fatal(err)
	}
	names, err := store.generations("discovery")
	if err != nil || len(names) != 2 {
		t.Fatal("rolling retention", names, err)
	}
	if err = os.WriteFile(filepath.Join(cache, names[0], "progress.bin"), []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.load("discovery"); err != nil {
		t.Fatal("intact previous generation not recovered", err)
	}
	if err = os.WriteFile(filepath.Join(cache, names[1], "progress.bin"), []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.load("discovery"); err == nil {
		t.Fatal("all corrupt generations accepted")
	}
}
