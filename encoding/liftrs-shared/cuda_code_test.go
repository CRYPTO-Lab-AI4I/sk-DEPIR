package liftrs

import (
	"bytes"
	"testing"
)

func TestCUDACodeResources(t *testing.T) {
	h, _, err := canonicalScalingCodeHeader(17, 5, 3876566203, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	for _, cards := range []int{1, 2, 4, 8} {
		r := cudaCodeResources(h, cards)
		if r.ArenaElements != h.N || r.PrefixElements != h.N/2 || r.CodeBytes != 4*h.K+8*h.N || r.HostOutputBytes != 4*h.N {
			t.Fatal("wide array layout", r)
		}
		if cards == 4 && r.MaxDeviceArrayBytes >= 112<<30 {
			t.Fatal("four-card principal arrays exceed intended cap", r)
		}
		if r.DeviceWorkspaceBytes < 16*h.N || r.DeviceWorkspaceBytes > 16*h.N+(32<<20) {
			t.Fatal("unexpected third workspace", r)
		}
	}
	p, raw := smallScalingCode(t)
	d, err := LoadScalingCode(bytes.NewReader(raw), codeDigest(raw), codeLimits(p))
	if err != nil {
		t.Fatal(err)
	}
	r, err := d.CUDAResources(4)
	if err != nil {
		t.Fatal(err)
	}
	if r.PrefixElements != 256 || r.ArenaElements != 512 {
		t.Fatal("dense prefix not admitted correctly", r)
	}
	for _, cards := range []int{0, 3, 16} {
		if _, err = d.CUDAResources(cards); err == nil {
			t.Fatal("invalid device count accepted")
		}
	}
	if _, err = d.WithCUDA(CUDACodeOptions{LibraryPath: "not-opened", Devices: []int{0, 0}, VectorLayout: "peer", DeviceMemoryLimit: 1 << 30}); err == nil {
		t.Fatal("duplicate device accepted")
	}
	if _, err = d.WithCUDA(CUDACodeOptions{LibraryPath: "not-opened", Devices: []int{0}, VectorLayout: "local", DeviceMemoryLimit: 1}); err == nil {
		t.Fatal("array cap not checked")
	}
}
