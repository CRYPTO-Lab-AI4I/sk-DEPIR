//go:build liftrs_cpp && cgo

package liftrscpp

/*
#include "c_api.h"
*/
import "C"

import "fmt"

// ClockReading observes independent clocks, in nanoseconds. SteadyNS uses the
// exact clock of native stage counters; ProcessCPUNS aggregates all threads.
// Compare differences, not absolute origins, across clock domains.
type ClockReading struct {
	SteadyNS     uint64 `json:"steady_ns"`
	WallNS       uint64 `json:"wall_ns"`
	MonotonicNS  uint64 `json:"posix_monotonic_ns"`
	ProcessCPUNS uint64 `json:"process_cpu_ns"`
}

// ReadClocks provides diagnostic readings without retaining any caller memory.
// Reading order introduces small skew, so this is not a nanosecond calibration.
func ReadClocks() (ClockReading, error) {
	var values [4]C.uint64_t
	var message [256]C.char
	if C.liftrs_read_clocks(&values[0], &message[0], C.size_t(len(message))) != 0 {
		return ClockReading{}, fmt.Errorf("native clocks: %s", C.GoString(&message[0]))
	}
	return ClockReading{uint64(values[0]), uint64(values[1]), uint64(values[2]), uint64(values[3])}, nil
}
