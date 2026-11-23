package liftrs

import (
	"encoding/json"
	"strings"
	"testing"
)

func commandFixture() DistributedPreprocessingConfig {
	return DistributedPreprocessingConfig{Spec: Spec{8, 2, 5}, Ranks: 4, Directory: "/owned/output", Library: "/trusted/lib.so", DeviceBytes: 64 << 20, MaxLocalWords: 16384, MaxGoodWords: 65536, MaxRootWords: 65536, RSSBytes: 1 << 30, StorageBytes: 16 << 20, FileBytes: 8 << 20, DeadlineSeconds: 240, NewtonInput: "prefix", Spectrum: "stream", DiscoveryResult: "stream", ProofLog: "trace"}
}
func TestDistributedJobAdmission(t *testing.T) {
	c := commandFixture()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	h, _, err := canonicalScalingCodeHeader(8, 5, 7381, 65536)
	if err != nil {
		t.Fatal(err)
	}
	if err = c.admitFiles(h); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(c)
	if _, err = ReadDistributedPreprocessingConfig(strings.NewReader(string(data))); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{string(data) + " {}", `{"unknown":1}`, strings.Repeat(" ", 65537) + string(data)} {
		if _, err = ReadDistributedPreprocessingConfig(strings.NewReader(body)); err == nil {
			t.Fatal("bad JSON accepted")
		}
	}
	for _, mutate := range []func(*DistributedPreprocessingConfig){func(v *DistributedPreprocessingConfig) { v.Ranks = 3 }, func(v *DistributedPreprocessingConfig) { v.DeviceBytes = 1 }, func(v *DistributedPreprocessingConfig) { v.NewtonInput = "" }, func(v *DistributedPreprocessingConfig) { v.MaxLocalWords-- }, func(v *DistributedPreprocessingConfig) { v.ResumeDirectory = v.Directory; v.Known = 8 }} {
		bad := c
		mutate(&bad)
		if err = bad.Validate(); err == nil {
			t.Fatal("bad job policy accepted")
		}
	}
	c.StorageBytes = 1
	if err = c.admitFiles(h); err == nil {
		t.Fatal("underbudget storage")
	}
	c = commandFixture()
	c.FileBytes = 1
	if err = c.admitFiles(h); err == nil {
		t.Fatal("underbudget canonical file")
	}
}
