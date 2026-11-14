//go:build liftrs_cpp && cgo

package liftrscpp

import "testing"

func TestPlanLocks(t *testing.T) {
	// A zero plan must fail without leaving a lock held, including when listed
	// repeatedly. No native pointer dereference is needed for these error cases.
	var plan Plan
	if _, err := lockPlans(&plan, &plan); err == nil {
		t.Fatal("closed plan")
	}
	plan.mu.Lock()
	plan.mu.Unlock()
	if _, err := lockPlans(nil); err == nil {
		t.Fatal("nil plan")
	}
	// Existing wide tests cover native lifetime concurrently; all old online and
	// disk tests now exercise the shared lock helper with real overlapping plans.
}
