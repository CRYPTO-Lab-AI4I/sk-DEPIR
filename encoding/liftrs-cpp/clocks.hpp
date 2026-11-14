#pragma once
// Internal C ABI implementation, included once by c_api.cpp.
#include "c_api.h"
#include <chrono>
#include <cstdio>
#include <ctime>
// Read the same steady_clock used by native stage timers, plus independent
// clocks for measurement diagnostics. These are observation only, not encoding.
extern "C" int liftrs_read_clocks(uint64_t *values, char *error, size_t capacity) {
  if (!values) {
    fail(error, capacity, "missing clock output");
    return 1;
  }
  timespec monotonic{}, cpu{};
  if (clock_gettime(CLOCK_MONOTONIC, &monotonic) ||
      clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &cpu)) {
    fail(error, capacity, "clock_gettime failed");
    return 1;
  }
  values[0] = std::chrono::duration_cast<std::chrono::nanoseconds>(
                  std::chrono::steady_clock::now().time_since_epoch())
                  .count();
  values[1] = std::chrono::duration_cast<std::chrono::nanoseconds>(
                  std::chrono::system_clock::now().time_since_epoch())
                  .count();
  values[2] = uint64_t(monotonic.tv_sec) * 1000000000 + monotonic.tv_nsec;
  values[3] = uint64_t(cpu.tv_sec) * 1000000000 + cpu.tv_nsec;
  return 0;
}
