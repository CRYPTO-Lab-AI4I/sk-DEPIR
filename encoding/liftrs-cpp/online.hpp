#pragma once

// The mathematical cyclic encoder, independent of word width and storage.
// Backends supply buffer access, transform execution and output representation.
#include "arithmetic.hpp"
#include "cyclic_schedule.hpp"
#include "matrix.hpp"
#include "progress.hpp"
#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <ctime>

namespace liftrs::execution {

// Chunk sizes remain backend choices, preserving each validated scheduling and
// workspace model. The reader can address resident or temporarily staged code.
template <class Word, class Read>
void pointwise(const FieldT<Word> &field, Word *data, size_t count, unsigned workers,
               size_t chunk, Read read) {
  parallel_jobs((count + chunk - 1) / chunk, workers, [&](size_t job) {
    for (size_t i = job * chunk; i < std::min(count, (job + 1) * chunk); ++i)
      data[i] = field.mul(data[i], Word(read(i)));
  });
}

// Scatter one contiguous coefficient range. start is the *global* coefficient
// offset, including for a disk tile. Primitive omega is checked before callers
// invoke this helper; distinct powers guarantee disjoint canonical positions.
// The caller owns parity storage so no memory strategy gains a hidden array.
template <class Word, class Read, class Put, class Reduce>
void scatter(const FieldT<Word> &field, Word omega, const ProductT<Word> &step,
             size_t start, size_t count, unsigned workers, size_t chunk, Read read,
             Put put, Reduce reduce, std::atomic<bool> &bad) {
  parallel_jobs((count + chunk - 1) / chunk, workers, [&](size_t job) {
    Word point = field_power(field, omega, start + job * chunk + 1);
    uint32_t sum = 0;
    for (size_t i = job * chunk; i < std::min(count, (job + 1) * chunk); ++i) {
      const auto value = read(i);
      if (uint64_t(value) >= (uint64_t(1) << field.s))
        bad.store(true, std::memory_order_relaxed);
      put(point, uint32_t(value));
      sum ^= uint32_t(value);
      point = step.mul(field, point);
    }
    reduce(job, sum);
  });
}

} // namespace liftrs::execution
