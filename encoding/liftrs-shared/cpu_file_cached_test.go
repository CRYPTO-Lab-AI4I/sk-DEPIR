//go:build liftrs_cpp && cgo

package liftrs

import (
	"os"
	"testing"
)

func TestCPUFileCached(t *testing.T) {
	out, cache := os.Getenv("PIR_CPU_FILE_OUT"), os.Getenv("PIR_CPU_FILE_CACHE")
	if out == "" {
		t.Skip("opt-in CPU file-backed encoding")
	}
	if cache == "" {
		t.Fatal("cache required")
	}
	spec := Spec{16, 2, 5}
	repeats := 1
	limit := int64(44 << 30)
	if os.Getenv("PIR_CPU_FILE_S14") == "1" {
		spec = Spec{14, 2, 5}
		repeats = 2
		limit = 8 << 30
	}
	if os.Getenv("PIR_CPU_FILE_SMOKE") == "1" {
		spec = Spec{6, 2, 4}
		repeats = 2
		limit = 1 << 30
	}
	streamed := os.Getenv("PIR_CPU_FILE_MODE") == "file"
	if !streamed {
		limit = 64 << 30
	}
	cfg := scalingFirstUseConfig{spec: spec, out: out, cache: cache, workers: 32, repeats: repeats, memoryLimit: limit, onlineOnly: true}
	var file *ScalingCPPFileEncoder
	if streamed {
		cfg.onlineLoad = func(path, digest string, max int) (*ScalingOnlinePlan, error) {
			var err error
			file, err = OpenScalingCPPFile(path, digest, max, 32)
			if err != nil {
				return nil, err
			}
			return file.cpu.source, nil
		}
	}
	cfg.onlineSetup = func(p *ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error) {
		if streamed {
			return func(input []uint32) ([]uint32, map[string]any, error) {
				before := file.WorkspaceBytes()
				word, st, err := file.EncodeWithStats(input)
				return word, map[string]any{"native_stages": st.ScalingCPPStats, "read_ns": st.ReadNS, "read_bytes": st.ReadBytes, "read_chunks": st.ReadChunks, "workspace_before_bytes": before, "workspace_after_bytes": file.WorkspaceBytes(), "file_streamed": true}, err
			}, func() { file.Close() }, map[string]any{"file_streamed": true, "workers": 32, "workspace_bytes": file.WorkspaceBytes()}, nil
		}
		e, err := p.WithCPPMatrix(32)
		if err != nil {
			return nil, nil, nil, err
		}
		w := e.NewWorkspace()
		return func(input []uint32) ([]uint32, map[string]any, error) {
			word, st, err := e.EncodeWithStats(input, w)
			return word, map[string]any{"native_stages": st, "file_streamed": false}, err
		}, func() { e.Close() }, map[string]any{"workers": 32, "file_streamed": false, "workspace_bytes": w.Bytes()}, nil
	}
	runScalingFirstUse(t, cfg)
}
