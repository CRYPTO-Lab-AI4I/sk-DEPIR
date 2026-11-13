//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func wideCodeFixture(t *testing.T, spec Spec) (*ScalingCodeData, *ScalingOnlinePlan) {
	t.Helper()
	basis, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(170058))})
	if err != nil {
		t.Fatal(err)
	}
	original, err := basis.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	// Full small-code polynomial identity P*G=1+X^(N-1), checked on all N
	// distinct field points using existing Go arithmetic before retaining data.
	check := make([]uint32, basis.Length())
	copy(check, original.p)
	basis.extensionFFT.apply(check, false)
	for j, x := range check {
		want := uint32(0)
		if j == 0 {
			want = 1
		}
		if basis.extension.mul(x, original.spectrum[j]) != want {
			t.Fatal("P*G identity", spec, j)
		}
	}
	plan, err := original.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	var raw bytes.Buffer
	if err = WriteScalingOnline(&raw, plan); err != nil {
		t.Fatal(err)
	}
	digest := codeDigest(raw.Bytes())
	d, err := LoadScalingCode(bytes.NewReader(raw.Bytes()), digest, ScalingCodeLimits{uint64(plan.Length()), plan.ResidentBytes()})
	if err != nil {
		t.Fatal(err)
	}
	if root := os.Getenv("PIR_WIDE_ENCODING_CACHE"); root != "" {
		dir := filepath.Join(root, fmt.Sprintf("s%d-m2-t%d", spec.FieldDegree, spec.CurveDegree))
		if err = os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, "online.bin"), raw.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		var generator bytes.Buffer
		if err = binary.Write(&generator, binary.LittleEndian, original.g); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, "g.bin"), generator.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		gHash := sha256.Sum256(generator.Bytes())
		meta := map[string]any{"s": spec.FieldDegree, "t": spec.CurveDegree, "k": plan.Dimension(), "n": plan.Length(), "bytes": raw.Len(), "sha256": digest, "identity_verified": true, "generator_sha256": hex.EncodeToString(gHash[:]), "seed": 170058}
		body, _ := json.MarshalIndent(meta, "", "  ")
		if err = os.WriteFile(filepath.Join(dir, "metadata.json"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return d, plan
}

func TestScalingWideCPPEncoding(t *testing.T) {
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {6, 2, 4}, {8, 2, 5}} {
		data, reference := wideCodeFixture(t, spec)
		var original bytes.Buffer
		WriteScalingCode(&original, data)
		message := make([]uint32, reference.Dimension())
		for j := range message {
			message[j] = uint32(j*7+3) & ((1 << uint(spec.FieldDegree)) - 1)
		}
		input := append([]uint32(nil), message...)
		want, err := reference.Encode(message, reference.NewWorkspace())
		if err != nil {
			t.Fatal(err)
		}
		// At small sizes add the independent dense code oracle, re-extracting its
		// own information set rather than assuming identical systematic order.
		if spec.FieldDegree <= 4 {
			oracle, err := Preprocess(spec)
			if err != nil {
				t.Fatal(err)
			}
			values := make([]uint32, oracle.Dimension())
			for j, p := range oracle.SystematicPositions() {
				values[j] = want[p]
			}
			dense, err := oracle.Encode(values)
			if err != nil || !reflect.DeepEqual(dense, want) {
				t.Fatal("dense oracle", err)
			}
		}
		for _, workers := range []int{1, 4} {
			encoder, err := data.WithWideCPP(workers)
			if err != nil {
				t.Fatal(err)
			}
			resources := encoder.Resources()
			if _, err = encoder.NewWorkspace(resources.WorkspaceBytes - 1); err == nil {
				t.Fatal("workspace cap ignored")
			}
			workspace, err := encoder.NewWorkspace(resources.WorkspaceBytes)
			if err != nil {
				t.Fatal(err)
			}
			first, stats, err := encoder.EncodeWithStats(message, workspace)
			if err != nil || !reflect.DeepEqual(first, want) {
				t.Fatal("wide/reference", spec, workers, err)
			}
			if cap(first) != len(first) || stats.Calls != 5 || stats.OutputBackingBytes != 8*uint64(reference.Length()) || workspace.Bytes()+stats.OutputBackingBytes != resources.WorkspaceBytes {
				t.Fatal("owned backing accounting")
			}
			for j, value := range message {
				point, err := data.SystematicPosition(uint64(j))
				if err != nil || first[point] != value {
					t.Fatal("systematic position", spec, j, err)
				}
			}
			saved := append([]uint32(nil), first...)
			for repeat := 0; repeat < 2; repeat++ {
				word, err := encoder.Encode(message, workspace)
				if err != nil || !reflect.DeepEqual(word, want) || !reflect.DeepEqual(first, saved) {
					t.Fatal("reuse overwrote prior output", err)
				}
			}
			zero, err := encoder.Encode(make([]uint32, len(message)), workspace)
			if err != nil || !reflect.DeepEqual(zero, make([]uint32, reference.Length())) {
				t.Fatal("zero encoding", err)
			}
			bad := append([]uint32(nil), message...)
			bad[0] = 1 << uint(spec.FieldDegree)
			if _, err = encoder.Encode(bad, workspace); err == nil {
				t.Fatal("bad message accepted")
			}
			if _, err = encoder.Encode(message[:len(message)-1], workspace); err == nil {
				t.Fatal("short message accepted")
			}
			recovered, err := encoder.Encode(message, workspace)
			if err != nil || !reflect.DeepEqual(recovered, want) || !reflect.DeepEqual(first, saved) || !reflect.DeepEqual(message, input) {
				t.Fatal("recovery/immutability", err)
			}
			var repacked bytes.Buffer
			var expected bytes.Buffer
			_, hash, err := writeCanonicalPacked(&repacked, first, spec.FieldDegree)
			if err != nil {
				t.Fatal(err)
			}
			_, wantHash, err := writeCanonicalPacked(&expected, want, spec.FieldDegree)
			if err != nil || hash != wantHash || !bytes.Equal(repacked.Bytes(), expected.Bytes()) {
				t.Fatal("packed identity", err)
			}
			encoder.Close()
			encoder.Close()
			runtime.GC()
			if encoder.OwnedBytes() != 0 || !reflect.DeepEqual(first, saved) {
				t.Fatal("close changed word")
			}
			if _, err = encoder.Encode(message, workspace); err != ErrCPPClosed {
				t.Fatal("closed encoder", err)
			}
		}
		var after bytes.Buffer
		WriteScalingCode(&after, data)
		if !bytes.Equal(original.Bytes(), after.Bytes()) {
			t.Fatal("code data changed")
		}
	}
}

func TestScalingWideCPPLifetime(t *testing.T) {
	data, _ := wideCodeFixture(t, Spec{3, 2, 2})
	e, err := data.WithWideCPP(2)
	if err != nil {
		t.Fatal(err)
	}
	message := make([]uint32, int(data.Header().K))
	for j := range message {
		message[j] = uint32(j % 8)
	}
	var group sync.WaitGroup
	for j := 0; j < 4; j++ {
		w, err := e.NewWorkspace(e.Resources().WorkspaceBytes)
		if err != nil {
			t.Fatal(err)
		}
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := e.Encode(message, w); err != nil && err != ErrCPPClosed {
				t.Error(err)
			}
		}()
	}
	e.Close()
	group.Wait()
	var empty ScalingWideCPPEncoder
	empty.Close()
	if empty.OwnedBytes() != 0 {
		t.Fatal("zero encoder")
	}
}
