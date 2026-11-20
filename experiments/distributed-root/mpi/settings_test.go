//go:build liftrs_mpi && cgo

package mpi

import "testing"

func TestTransportSettingsSnapshot(t *testing.T) {
	t.Setenv("UCX_RCACHE_ENABLE", "n")
	t.Setenv("PIR_DIAGNOSTIC_SECRET", "excluded")
	a := TransportSettings()
	if a["UCX_RCACHE_ENABLE"] != "n" || len(a) != 6 {
		t.Fatal("control missing or allowlist changed")
	}
	if _, ok := a["PIR_DIAGNOSTIC_SECRET"]; ok {
		t.Fatal("unrelated environment exposed")
	}
	a["UCX_RCACHE_ENABLE"] = "try"
	if TransportSettings()["UCX_RCACHE_ENABLE"] != "n" {
		t.Fatal("snapshot aliases shared state")
	}
}
