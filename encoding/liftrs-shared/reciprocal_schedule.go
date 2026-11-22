package liftrs

import "fmt"

// reciprocalSchedule shares completed-prefix doubling between host and sharded
// Newton implementations. A failed refinement never advances the prefix.
func reciprocalSchedule(known, target int, refine func(int) error) error {
	if known < 1 || refine == nil {
		return fmt.Errorf("reciprocal prefix schedule")
	}
	for known < target {
		size := target
		if known <= target/2 {
			size = 2 * known
		}
		if err := refine(size); err != nil {
			return err
		}
		known = size
	}
	return nil
}
