#pragma once

// Bounded little-endian word I/O used by streamed code and disk transforms.
// These counters are logical calls/bytes, not physical device traffic. Callers
// retain their existing overlapping phase-clock and chunk-count conventions.
#include <algorithm>
#include <cerrno>
#include <chrono>
#include <cstdint>
#include <limits>
#include <stdexcept>
#include <unistd.h>

namespace liftrs::io {
struct Counters {
  uint64_t ns = 0, bytes = 0, calls = 0;
};

inline uint64_t elapsed(std::chrono::steady_clock::time_point start) {
  return std::chrono::duration_cast<std::chrono::nanoseconds>(
             std::chrono::steady_clock::now() - start)
      .count();
}

inline void check_extent(uint64_t offset, size_t count) {
  static_assert(sizeof(off_t) >= 8, "file encoding requires wide offsets");
  if (offset > uint64_t(INT64_MAX) ||
      uint64_t(count) > (uint64_t(INT64_MAX) - offset) / 4)
    throw std::runtime_error("file offset overflow");
}

inline Counters read_words(int fd, uint64_t offset, uint32_t *output, size_t count,
                           uint64_t limit,
                           size_t max_read = std::numeric_limits<size_t>::max()) {
  check_extent(offset, count);
  const auto start = std::chrono::steady_clock::now();
  Counters counters;
  auto *bytes = reinterpret_cast<unsigned char *>(output);
  const size_t total = count * 4;
  size_t done = 0;
  while (done < total) {
    const auto got = ::pread(fd, bytes + done, std::min(total - done, max_read),
                             off_t(offset + done));
    ++counters.calls;
    if (got < 0 && errno == EINTR)
      continue;
    if (got <= 0)
      throw std::runtime_error("code file pread failed/truncated");
    done += size_t(got);
    counters.bytes += uint64_t(got);
  }
  for (size_t i = 0; i < count; ++i) {
    const size_t j = 4 * i;
    output[i] = uint32_t(bytes[j]) | (uint32_t(bytes[j + 1]) << 8) |
                (uint32_t(bytes[j + 2]) << 16) | (uint32_t(bytes[j + 3]) << 24);
    if (uint64_t(output[i]) >= limit)
      throw std::runtime_error("file code element outside field");
  }
  counters.ns = elapsed(start);
  return counters;
}

inline Counters write_words(int fd, uint64_t offset, const uint32_t *input,
                            size_t count) {
  check_extent(offset, count);
  const auto start = std::chrono::steady_clock::now();
  Counters counters;
  unsigned char bytes[65536];
  for (size_t first = 0; first < count;) {
    const size_t words = std::min(count - first, size_t(16384));
    for (size_t i = 0; i < words; ++i)
      for (unsigned j = 0; j < 4; ++j)
        bytes[4 * i + j] = static_cast<unsigned char>(input[first + i] >> (8 * j));
    size_t done = 0;
    while (done < words * 4) {
      const auto sent = ::pwrite(fd, bytes + done, words * 4 - done,
                                 off_t(offset + first * 4 + done));
      ++counters.calls;
      if (sent < 0 && errno == EINTR)
        continue;
      if (sent <= 0)
        throw std::runtime_error("code file pwrite failed");
      done += size_t(sent);
      counters.bytes += uint64_t(sent);
    }
    first += words;
  }
  counters.ns = elapsed(start);
  return counters;
}
} // namespace liftrs::io
