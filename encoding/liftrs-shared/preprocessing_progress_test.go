package liftrs

import "testing"

// Dense synthetic exponents test validation/chunk boundaries, not Lift-RS
// dimension. q permutes [0,N-2] under Frobenius, so the whole set is closed.
func TestSupportValidationProgress(t *testing.T) {
	const n = 1 << 18
	support := make([]int, n-1)
	for i := range support {
		support[i] = i
	}
	progress := map[string][]uint64{}
	notice := func(phase string, done, total uint64, unit string) {
		if total != n-1 || unit != "exponents" {
			t.Fatal("progress extent/units", total, unit)
		}
		prior := progress[phase]
		if len(prior) > 0 && done <= prior[len(prior)-1] {
			t.Fatal("nonadvancing progress")
		}
		progress[phase] = append(prior, done)
	}
	if err := validateExponentSupport(support, n, 1<<9, notice); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"support-order-validation", "support-Frobenius-validation"} {
		rows := progress[phase]
		if len(rows) != 4 || rows[0] != 65536 || rows[3] != n-1 {
			t.Fatal("missing chunk progress", phase, rows)
		}
	}
	support[65536] = support[65535]
	if err := validateExponentSupport(support, n, 1<<9); err == nil {
		t.Fatal("cross-chunk duplicate accepted")
	}
}
