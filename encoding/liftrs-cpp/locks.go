//go:build liftrs_cpp && cgo

package liftrscpp

import "errors"

// A call uses at most four immutable plans. Fixed storage avoids a map and
// prevents repeated RLock on shared prefix/full handles: a pending Close would
// otherwise deadlock recursive read locking. Release in reverse acquisition order.
type planLocks struct {
	plans [4]*Plan
	count int
}

func lockPlans(plans ...*Plan) (planLocks, error) {
	var held planLocks
	for _, plan := range plans {
		if plan == nil {
			held.release()
			return planLocks{}, errors.New("nil native plan")
		}
		seen := false
		for _, existing := range held.plans[:held.count] {
			if existing == plan {
				seen = true
				break
			}
		}
		if seen {
			continue
		}
		if held.count == len(held.plans) {
			held.release()
			return planLocks{}, errors.New("too many native plans")
		}
		plan.mu.RLock()
		held.plans[held.count] = plan
		held.count++
		if plan.handle == nil {
			held.release()
			return planLocks{}, errors.New("native plan closed")
		}
	}
	return held, nil
}

func (held *planLocks) release() {
	for held.count > 0 {
		held.count--
		held.plans[held.count].mu.RUnlock()
	}
}
