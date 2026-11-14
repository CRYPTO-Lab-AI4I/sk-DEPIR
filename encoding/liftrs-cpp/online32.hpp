#pragma once
// Internal C ABI implementation, included once by c_api.cpp.
#include "c_api.h"
#include "file_io.hpp"
#include "online.hpp"
#include "output.hpp"
#include "plan.hpp"
#include <cstdio>
#include <cstring>

static int online_encode(bool compact, size_t scratch_count, int code_fd,
                         uint64_t p_offset, uint64_t spectrum_offset, size_t tile_count,
                         const liftrs_plan *pr, const liftrs_plan *pc,
                         const liftrs_plan *fr, const liftrs_plan *fc, unsigned ps,
                         unsigned fs, unsigned workers, uint32_t omega,
                         const uint32_t *message, const uint32_t *p,
                         const uint32_t *spectrum, size_t k, size_t n, size_t l,
                         uint32_t *arena, size_t arena_count, uint32_t *h,
                         uint32_t *scratch, uint32_t *word, uint64_t *stats,
                         char *error, size_t capacity, bool packed = false) {
  try {
    if ((packed && (compact || code_fd < 0)) || !pr || !pc || !fr || !fc || !message ||
        (code_fd < 0 && (!p || !spectrum)) || !arena ||
        (!compact && code_fd < 0 && !h) || !scratch || !word || !stats || !k ||
        k > n / 2 || !l || (l & (l - 1)) || l < 2 * k - 1 || l > n ||
        arena_count < std::max(2 * l, n) || ((compact || code_fd >= 0) && l > n / 2) ||
        (compact && word != arena) || !workers || workers > 64 ||
        fr->transform.f.q != n) {
      throw std::runtime_error("invalid online shape/buffers");
    }
    const auto &f = fr->transform.f;
    auto same = [&](const liftrs_plan *a) {
      return a->transform.f.q == f.q && a->transform.f.s == f.s &&
             a->transform.f.polynomial == f.polynomial &&
             a->transform.f.delta == f.delta;
    };
    if (!same(pr) || !same(pc) || !same(fc)) {
      throw std::runtime_error("online field mismatch");
    }
    if (!liftrs::is_primitive(f, omega))
      throw std::runtime_error("nonprimitive online generator");
    for (size_t i = 0; i < k; i++) {
      if (message[i] >= (1u << f.s) || (code_fd < 0 && p[i] >= (1u << f.s))) {
        throw std::runtime_error("online input outside base field");
      }
    }
    if (code_fd >= 0 && (p_offset > uint64_t(INT64_MAX) - uint64_t(k + 1) * 4 ||
                         spectrum_offset > uint64_t(INT64_MAX) - uint64_t(n) * 4))
      throw std::runtime_error("file offset overflow");
    if (code_fd >= 0 && (compact || word != scratch || !tile_count ||
                         tile_count > std::min(n, size_t(1) << 26)))
      throw std::runtime_error("file streaming shape/tile");
    std::fill(stats, stats + (code_fd >= 0 ? 11 : compact ? 8 : 6), 0);
    auto measured = [&](unsigned stage, auto fn) {
      auto start = std::chrono::steady_clock::now();
      fn();
      stats[stage] += std::chrono::duration_cast<std::chrono::nanoseconds>(
                          std::chrono::steady_clock::now() - start)
                          .count();
    };
    auto read_words = [&](uint32_t *destination, size_t count, uint64_t offset,
                          uint64_t limit) {
      auto io = liftrs::io::read_words(code_fd, offset, destination, count, limit,
                                       size_t(1) << 22);
      stats[8] += io.ns;
      stats[9] += io.bytes;
      stats[10]++; // Logical chunks, unlike disk syscall counters.
    };
    auto fft = [&](bool prefix, uint32_t *data, bool inverse) {
      uint64_t stages[3] = {};
      liftrs::execution::matrix(compact, scratch_count, prefix ? pr : fr,
                                prefix ? pc : fc, data, scratch, prefix ? l : n,
                                prefix ? ps : fs, workers, inverse, stages);
      for (unsigned j = 0; j < 3; j++)
        stats[j] += stages[j];
    };
    auto *x = arena;
    auto *z = arena + l;
    auto prepare = [&] {
      measured(3, [&] {
        std::fill(arena, arena + 2 * l, 0);
        if (code_fd >= 0) {
          for (size_t first = 0; first < k + 1; first += tile_count) {
            size_t count = std::min(tile_count, k + 1 - first);
            read_words(x + first, count, p_offset + first * 4, uint64_t(1) << f.s);
          }
          if (x[0] != 1)
            throw std::runtime_error("file P normalization");
          x[k] = 0; // P[K] is verified but excluded by the existing prefix algorithm.
        } else {
          std::copy(p, p + k, x);
        }
        std::copy(message, message + k, z);
      });
    };
    constexpr size_t chunk = 65536;
    auto multiply = [&](uint32_t *a, const uint32_t *b, size_t count) {
      liftrs::execution::pointwise(f, a, count, workers, chunk,
                                   [&](size_t i) { return b[i]; });
    };
    auto prefix_product = [&] { measured(4, [&] { multiply(x, z, l); }); };
    auto promote = [&] {
      measured(3, [&] {
        if (compact || code_fd >= 0) {
          std::fill(arena + k, arena + n, 0);
        } else {
          std::copy(x, x + k, h);
          std::fill(arena, arena + n, 0);
          std::copy(h, h + k, arena);
        }
      });
    };
    auto spectrum_product = [&] {
      if (code_fd >= 0) {
        // Full forward has joined. Scratch is dead until inverse, so the file tile
        // needs no allocation. Each product joins before the next read reuses it.
        for (size_t first = 0; first < n; first += tile_count) {
          size_t count = std::min(tile_count, n - first);
          read_words(scratch, count, spectrum_offset + first * 4, f.q);
          measured(4, [&] { multiply(arena + first, scratch, count); });
        }
      } else {
        measured(4, [&] { multiply(arena, spectrum, n); });
      }
    };
    auto assemble = [&] {
      auto *bytes = reinterpret_cast<unsigned char *>(arena);
      if (compact)
        measured(6, [&] { liftrs::output::narrow_inplace(arena, n, f.s); });
      measured(5, [&] {
        // One small fixed-constant lookup map replaces scalar multiplication by
        // omega. Each chunk starts at omega^(first+1), so no N-sized map is needed.
        // Distinct field coordinates can share packed bytes. Atomic byte OR
        // joins disjoint bit fields without endian-dependent word stores.
        auto *packed_bytes = reinterpret_cast<unsigned char *>(word);
        auto put_packed = [&](size_t point, uint32_t value) {
          liftrs::output::or_packed(packed_bytes, point, value, f.s);
        };
        if (packed)
          std::memset(packed_bytes, 0, (n * f.s + 7) / 8);
        const auto step = liftrs::constant_product(f, omega);
        const size_t jobs = (n - 1 + chunk - 1) / chunk;
        std::vector<uint32_t> parity(jobs);
        std::atomic<bool> bad{false};
        liftrs::execution::scatter(
            f, omega, step, 0, n - 1, workers, chunk,
            [&](size_t i) {
              return compact ? liftrs::output::load_le_u16(bytes, i) : arena[i];
            },
            [&](uint32_t point, uint32_t value) {
              if (packed) {
                if (value < (1u << f.s))
                  put_packed(point, value);
              } else if (compact) {
                liftrs::output::store_le_u16(bytes + 2 * n, size_t(point), value);
              } else
                word[point] = value;
            },
            [&](size_t job, uint32_t sum) { parity[job] = sum; }, bad);
        if (bad.load())
          throw std::runtime_error("product escaped base field");
        uint32_t zero = 0;
        for (auto value : parity)
          zero ^= value;
        if (packed) {
          put_packed(0, zero);
        } else if (compact) {
          bytes[2 * n] = static_cast<unsigned char>(zero);
          bytes[2 * n + 1] = static_cast<unsigned char>(zero >> 8);
        } else {
          word[0] = zero;
        }
      });
      if (compact) {
        measured(6, [&] { std::memcpy(bytes, bytes + 2 * n, 2 * n); });
        measured(7, [&] { liftrs::output::expand_inplace(arena, n); });
      }
    };
    liftrs::execution::encode_cyclic(
        prepare,
        [&](liftrs::execution::Input input, bool inverse) {
          fft(input != liftrs::execution::Input::full,
              input == liftrs::execution::Input::prefix_message ? z : x, inverse);
        },
        prefix_product, promote, spectrum_product, assemble);
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "native online encoding failed");
    return 1;
  }
}

extern "C" int liftrs_online_encode(
    const liftrs_plan *pr, const liftrs_plan *pc, const liftrs_plan *fr,
    const liftrs_plan *fc, unsigned ps, unsigned fs, unsigned workers, uint32_t omega,
    const uint32_t *message, const uint32_t *p, const uint32_t *spectrum, size_t k,
    size_t n, size_t l, uint32_t *arena, size_t arena_count, uint32_t *h,
    uint32_t *scratch, uint32_t *word, uint64_t *stats, char *error, size_t capacity) {
  return online_encode(false, n, -1, 0, 0, 0, pr, pc, fr, fc, ps, fs, workers, omega,
                       message, p, spectrum, k, n, l, arena, arena_count, h, scratch,
                       word, stats, error, capacity);
}
extern "C" int liftrs_online_encode_compact(
    size_t scratch_count, const liftrs_plan *pr, const liftrs_plan *pc,
    const liftrs_plan *fr, const liftrs_plan *fc, unsigned ps, unsigned fs,
    unsigned workers, uint32_t omega, const uint32_t *message, const uint32_t *p,
    const uint32_t *spectrum, size_t k, size_t n, size_t l, uint32_t *arena,
    size_t arena_count, uint32_t *h, uint32_t *scratch, uint32_t *word, uint64_t *stats,
    char *error, size_t capacity) {
  return online_encode(true, scratch_count, -1, 0, 0, 0, pr, pc, fr, fc, ps, fs,
                       workers, omega, message, p, spectrum, k, n, l, arena,
                       arena_count, h, scratch, word, stats, error, capacity);
}

// Explicit file-backed fast schedule: stats[11] add read/range-check ns, bytes,
// logical chunks. P reads are also inside buffer time; do not sum that twice.
extern "C" int liftrs_online_encode_file(
    int fd, uint64_t p_offset, uint64_t spectrum_offset, size_t tile,
    const liftrs_plan *pr, const liftrs_plan *pc, const liftrs_plan *fr,
    const liftrs_plan *fc, unsigned ps, unsigned fs, unsigned workers, uint32_t omega,
    const uint32_t *message, size_t k, size_t n, size_t l, uint32_t *arena,
    uint32_t *scratch, uint64_t *stats, char *error, size_t capacity) {
  if (fd < 0) {
    fail(error, capacity, "invalid code descriptor");
    return 1;
  }
  return online_encode(false, n, fd, p_offset, spectrum_offset, tile, pr, pc, fr, fc,
                       ps, fs, workers, omega, message, nullptr, nullptr, k, n, l,
                       arena, n, nullptr, scratch, scratch, stats, error, capacity);
}

extern "C" int liftrs_online_encode_file_packed(
    int fd, uint64_t p_offset, uint64_t spectrum_offset, size_t tile,
    const liftrs_plan *pr, const liftrs_plan *pc, const liftrs_plan *fr,
    const liftrs_plan *fc, unsigned ps, unsigned fs, unsigned workers, uint32_t omega,
    const uint32_t *message, size_t k, size_t n, size_t l, uint32_t *arena,
    uint32_t *scratch, uint64_t *stats, char *error, size_t capacity) {
  return online_encode(false, n, fd, p_offset, spectrum_offset, tile, pr, pc, fr, fc,
                       ps, fs, workers, omega, message, nullptr, nullptr, k, n, l,
                       arena, n, nullptr, scratch, scratch, stats, error, capacity,
                       true);
}
