#pragma once
// Internal C ABI implementation, included once by c_api.cpp.
#include "c_api.h"
#include "file_io.hpp"
#include "matrix.hpp"
#include "online.hpp"
#include "plan.hpp"
#include <cstdio>
#include <fcntl.h>
#include <sys/stat.h>

// Complete explicit disk schedule. The output alone stays in caller RAM;
// code and mutable domains stay in files. All field arithmetic is native.
extern "C" int liftrs_online_encode_disk(
    const liftrs_plan *pr, const liftrs_plan *pc, const liftrs_plan *fr,
    const liftrs_plan *fc, unsigned ps, unsigned fs, unsigned workers, uint32_t omega,
    int code_fd, uint64_t p_offset, uint64_t spectrum_offset, int data_fd,
    int scratch_fd, size_t tile, const uint32_t *message, size_t k, size_t n, size_t l,
    uint32_t *word, uint64_t *stats, char *error, size_t capacity) {
  try {
    if (!pr || !pc || !fr || !fc || !message || !word || !stats || !k || k > n / 2 ||
        l < 2 * k - 1 || l > n / 2 || (l & (l - 1)) || tile < 4 ||
        tile > (size_t(1) << 26) || !workers || workers > 64 || code_fd < 0 ||
        data_fd < 0 || scratch_fd < 0 || n != fr->transform.f.q)
      throw std::runtime_error("disk online arguments");
    const auto &f = fr->transform.f;
    for (auto *plan : {pr, pc, fc})
      if (plan->transform.f.q != f.q || plan->transform.f.s != f.s ||
          plan->transform.f.polynomial != f.polynomial ||
          plan->transform.f.delta != f.delta)
        throw std::runtime_error("disk online fields");
    struct stat files[3];
    int descriptors[3] = {code_fd, data_fd, scratch_fd};
    for (unsigned i = 0; i < 3; i++) {
      if (fstat(descriptors[i], files + i) || !S_ISREG(files[i].st_mode))
        throw std::runtime_error("disk online regular files");
      for (unsigned j = 0; j < i; j++)
        if (files[i].st_dev == files[j].st_dev && files[i].st_ino == files[j].st_ino)
          throw std::runtime_error("disk online file alias");
      int flags = fcntl(descriptors[i], F_GETFL);
      if (flags < 0 ||
          (i ? ((flags & O_ACCMODE) != O_RDWR) : ((flags & O_ACCMODE) == O_WRONLY)))
        throw std::runtime_error("disk online descriptor access");
    }
    // Probe headers are25/33 bytes; code-word ranges need not be aligned.
    if (p_offset > uint64_t(INT64_MAX) - 4 * uint64_t(k + 1) ||
        spectrum_offset > uint64_t(INT64_MAX) - 4 * uint64_t(n) ||
        uint64_t(files[0].st_size) < std::max(p_offset + 4 * uint64_t(k + 1),
                                              spectrum_offset + 4 * uint64_t(n)) ||
        uint64_t(files[1].st_size) < 4 * uint64_t(n) ||
        uint64_t(files[2].st_size) < 4 * uint64_t(n))
      throw std::runtime_error("disk online file length/offset");
    if (!liftrs::is_primitive(f, omega))
      throw std::runtime_error("disk online nonprimitive generator");
    for (size_t i = 0; i < k; i++)
      if (message[i] >= (1u << f.s))
        throw std::runtime_error("disk online message outside field");
    std::fill(stats, stats + 16, 0);
    auto begin = std::chrono::steady_clock::now();
    tile = std::min(tile, n);
    std::vector<uint32_t> a(tile), b(tile);
    if (a.capacity() > tile || b.capacity() > tile)
      throw std::runtime_error("disk online tile capacity");
    // Outer product/assembly tiles remain live during nested transforms, which
    // own two more tiles. Include both pairs and conservative I/O stack staging.
    stats[12] = 16 * uint64_t(tile) + 2 * 65536;
    stats[15] = 8 * uint64_t(n);
    auto ns = [](auto start) {
      return uint64_t(std::chrono::duration_cast<std::chrono::nanoseconds>(
                          std::chrono::steady_clock::now() - start)
                          .count());
    };
    auto read = [&](int fd, uint64_t offset, uint32_t *dst, size_t count,
                    uint64_t limit) {
      auto io = liftrs::io::read_words(fd, offset, dst, count, limit);
      stats[6] += io.ns;
      stats[8] += io.bytes;
      stats[10] += io.calls;
    };
    auto write = [&](uint64_t offset, const uint32_t *src, size_t count) {
      auto io = liftrs::io::write_words(data_fd, offset, src, count);
      stats[7] += io.ns;
      stats[9] += io.bytes;
      stats[11] += io.calls;
    };
    auto measured = [&](unsigned phase, auto fn) {
      auto start = std::chrono::steady_clock::now();
      fn();
      stats[phase] += ns(start);
    };
    auto zero = [&](size_t first, size_t count) {
      std::fill(a.begin(), a.end(), 0);
      liftrs::execution::WorkProgress progress("disk-zero-words", count);
      while (count) {
        size_t take = std::min(tile, count);
        write(4 * uint64_t(first), a.data(), take);
        first += take;
        count -= take;
        progress.advance(take);
      }
    };
    auto fft = [&](bool prefix, uint64_t offset, bool inverse) {
      uint64_t st[12];
      char detail[512];
      if (liftrs_matrix_apply_file(prefix ? pr : fr, prefix ? pc : fc, data_fd,
                                   scratch_fd, offset, 0, prefix ? l : n, tile,
                                   prefix ? ps : fs, workers, inverse, st, detail,
                                   sizeof(detail)))
        throw std::runtime_error(detail);
      for (unsigned i = 0; i < 3; i++)
        stats[i] += st[i];
      for (unsigned i = 3; i < 9; i++)
        stats[i + 3] += st[i];
      stats[14]++;
    };
    auto product = [&](bool prefix) {
      size_t count = prefix ? l : n;
      liftrs::execution::WorkProgress progress("disk-product-tiles",
                                               (count + tile - 1) / tile);
      for (size_t first = 0; first < count;) {
        size_t take = std::min(tile, count - first);
        read(data_fd, 4 * uint64_t(first), a.data(), take, f.q);
        read(prefix ? data_fd : code_fd,
             (prefix ? 4 * uint64_t(l) : spectrum_offset) + 4 * uint64_t(first),
             b.data(), take, f.q);
        measured(4, [&] {
          liftrs::execution::pointwise(f, a.data(), take, workers, 8192,
                                       [&](size_t i) { return b[i]; });
        });
        write(4 * uint64_t(first), a.data(), take);
        first += take;
        progress.advance();
      }
    };
    auto prepare = [&] {
      measured(3, [&] {
        zero(0, 2 * l);
        for (size_t first = 0; first < k;) {
          size_t take = std::min(tile, k - first);
          read(code_fd, p_offset + 4 * uint64_t(first), a.data(), take,
               uint64_t(1) << f.s);
          if (!first && a[0] != 1)
            throw std::runtime_error("disk P normalization");
          write(4 * uint64_t(first), a.data(), take);
          first += take;
        }
        read(code_fd, p_offset + 4 * uint64_t(k), b.data(), 1, uint64_t(1) << f.s);
        for (size_t first = 0; first < k;) {
          size_t take = std::min(tile, k - first);
          write(4 * uint64_t(l + first), message + first, take);
          first += take;
        }
      });
    };
    auto promote = [&] { measured(3, [&] { zero(k, n - k); }); };
    auto assemble = [&] {
      const auto step = liftrs::constant_product(f, omega);
      uint32_t sum = 0;
      liftrs::execution::WorkProgress progress("disk-assembly-coefficients", n - 1);
      for (size_t first = 0; first < n - 1;) {
        size_t take = std::min(tile, n - 1 - first);
        read(data_fd, 4 * uint64_t(first), a.data(), take, f.q);
        measured(5, [&] {
          std::atomic<uint32_t> parity{0};
          std::atomic<bool> bad{false};
          constexpr size_t chunk = 4096;
          liftrs::execution::scatter(
              f, omega, step, first, take, workers, chunk,
              [&](size_t i) { return a[i]; },
              [&](uint32_t point, uint32_t value) { word[point] = value; },
              [&](size_t, uint32_t local) { parity.fetch_xor(local); }, bad);
          if (bad.load())
            throw std::runtime_error("disk result outside base field");
          sum ^= parity.load();
        });
        first += take;
        progress.advance(take);
      }
      word[0] = sum;
    };
    liftrs::execution::encode_cyclic(
        prepare,
        [&](liftrs::execution::Input input, bool inverse) {
          fft(input != liftrs::execution::Input::full,
              input == liftrs::execution::Input::prefix_message ? 4 * uint64_t(l) : 0,
              inverse);
        },
        [&] { product(true); }, promote, [&] { product(false); }, assemble);
    stats[13] = ns(begin);
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "disk online encoding failed");
    return 1;
  }
}
