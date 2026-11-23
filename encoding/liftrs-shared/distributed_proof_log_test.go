package liftrs

import (
	"fmt"
	"testing"
)

func TestProofLogOrderingAndBounds(t *testing.T) {
	trace, _ := newProofLog("trace")
	summary, _ := newProofLog("summary")
	for j := 0; j < 100; j++ {
		row := map[string]any{"kind": "operand-proof", "case": "s8/newton-8", "packet": operandPacket{1, 1 << 61, 0, 1 << 60, 4, 0, 1}, "own_proof": fmt.Sprintf("%064x", j)}
		if hide, err := trace.Record(row); err != nil || hide {
			t.Fatal(hide, err)
		}
		row["utc"] = "ignored"
		row["rank"] = 99
		row["elapsed_seconds"] = 1.2
		if hide, err := summary.Record(row); err != nil || !hide {
			t.Fatal(hide, err)
		}
	}
	a, b := trace.Summaries(true), summary.Summaries(true)
	if len(a) != 1 || len(b) != 1 || a[0]["sha256"] != b[0]["sha256"] || a[0]["records"] != uint64(100) || len(summary.counts) != 1 {
		t.Fatal("proof summary")
	}
	if hide, err := summary.Record(map[string]any{"kind": "progress"}); err != nil || hide {
		t.Fatal("nonproof suppressed")
	}
	if summary.Summaries(false)[0]["complete"] != false {
		t.Fatal("partial summary certified")
	}
	reverse, _ := newProofLog("summary")
	for j := 99; j >= 0; j-- {
		reverse.Record(map[string]any{"kind": "operand-proof", "case": "s8/newton-8", "packet": operandPacket{1, 1 << 61, 0, 1 << 60, 4, 0, 1}, "own_proof": fmt.Sprintf("%064x", j)})
	}
	if reverse.Summaries(true)[0]["sha256"] == a[0]["sha256"] {
		t.Fatal("ordering lost")
	}
}

func TestProofLogAdmissionAndSerialization(t *testing.T) {
	for _, mode := range []string{"", "full"} {
		p, err := newProofLog(mode)
		if err != nil || p != nil {
			t.Fatal(mode, err)
		}
		if hide, err := p.Record(map[string]any{"kind": "tile-proof"}); err != nil || hide || len(p.Summaries(true)) != 0 {
			t.Fatal("default changed")
		}
	}
	if _, err := newProofLog("unknown"); err == nil {
		t.Fatal("unknown mode")
	}
	p, _ := newProofLog("summary")
	if _, err := p.Record(map[string]any{"kind": "operand-proof", "bad": make(chan int)}); err == nil || p.total != 0 {
		t.Fatal("bad evidence accepted")
	}
}

// The independent Python reviewer uses this same exact semantic fixture.
func TestProofLogGolden(t *testing.T) {
	p, _ := newProofLog("summary")
	p.Record(map[string]any{"kind": "operand-proof", "packet": operandPacket{1, 1 << 61, 0, 1 << 60, 4, 0, 1}, "case": "golden"})
	const expected = "6cbcbc0ee427991c40a2720d541b47ffe8ad8d6cfb3a32c23dcf4da760843776"
	if p.Summaries(true)[0]["sha256"] != expected {
		t.Fatal(p.Summaries(true))
	}
}
