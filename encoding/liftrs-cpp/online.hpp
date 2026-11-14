#pragma once

// The mathematical cyclic encoder, independent of word width and storage.
// Backends supply buffer access, transform execution and output representation.
#include "arithmetic.hpp"
#include "matrix.hpp"
#include "progress.hpp"
#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <ctime>

namespace liftrs::execution {

enum class Input { prefix_p, prefix_message, full };

// H=(P*M) mod X^K, Y=G*H, then phi(Y) and parity at coordinate zero.
// Every callback joins its workers before returning. That barrier is essential:
// transforms share scratch, and promotion deliberately reuses the prefix arena.
// Keeping this order here prevents resident/wide/file/disk implementations from
// drifting apart. Callbacks are inlined; no virtual calls or domain allocations.
template <class Prepare, class Transform, class PrefixProduct, class Promote,
          class SpectrumProduct, class Assemble>
void encode_cyclic(Prepare prepare, Transform transform, PrefixProduct prefix_product,
                   Promote promote, SpectrumProduct spectrum_product,
                   Assemble assemble) {
  EncodingProgress progress;
  progress.run("prepare-prefix", prepare);
  progress.run("prefix-p-forward", [&] { transform(Input::prefix_p, false); });
  progress.run("prefix-message-forward",
               [&] { transform(Input::prefix_message, false); });
  progress.run("prefix-product", prefix_product);
  progress.run("prefix-inverse", [&] { transform(Input::prefix_p, true); });
  progress.run("promote-prefix", promote);
  progress.run("full-forward", [&] { transform(Input::full, false); });
  progress.run("spectrum-product", spectrum_product);
  progress.run("full-inverse", [&] { transform(Input::full, true); });
  progress.run("canonical-assembly", assemble);
}

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
