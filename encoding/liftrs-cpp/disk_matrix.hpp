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

extern "C" int liftrs_matrix_apply_file(const liftrs_plan *rows,
                                        const liftrs_plan *columns, int data_fd,
                                        int scratch_fd, uint64_t data_offset,
                                        uint64_t scratch_offset, size_t n, size_t tile,
                                        unsigned split, unsigned workers, int inverse,
                                        uint64_t *stats, char *error, size_t capacity) {
  try {
    if (!rows || !columns || !stats || data_fd < 0 || scratch_fd < 0 || !workers ||
        workers > 64 || (inverse != 0 && inverse != 1) || n < 4 || (n & (n - 1)) ||
        tile < 4 || tile > (size_t(1) << 26))
      throw std::runtime_error("disk transform arguments");
    const auto &r = rows->transform;
    const auto &c = columns->transform;
    unsigned dim = 0;
    for (size_t v = n; v > 1; v >>= 1)
      dim++;
    if (!split || split >= dim || n > r.f.q || r.terms.size() != split + 1 ||
        c.terms.size() != dim - split || rows->taylor_terms.size() != dim - split ||
        r.f.q != c.f.q || r.f.polynomial != c.f.polynomial || r.f.delta != c.f.delta)
      throw std::runtime_error("disk transform shape/field");
    size_t n1 = size_t(1) << split, n2 = n / n1;
    if (tile < std::max(n1, n2))
      throw std::runtime_error("disk tile shorter than vector");
    for (size_t i = 0; i < r.basis.size(); i++)
      if (r.basis[i] != (uint32_t(1) << i))
        throw std::runtime_error("disk rows require canonical basis");
    struct stat data_info, scratch_info;
    if (fstat(data_fd, &data_info) || fstat(scratch_fd, &scratch_info) ||
        !S_ISREG(data_info.st_mode) || !S_ISREG(scratch_info.st_mode) ||
        (data_info.st_dev == scratch_info.st_dev &&
         data_info.st_ino == scratch_info.st_ino))
      throw std::runtime_error("disk files must be distinct regular files");
    for (auto fd : {data_fd, scratch_fd}) {
      int flags = fcntl(fd, F_GETFL);
      if (flags < 0 || (flags & O_ACCMODE) != O_RDWR)
        throw std::runtime_error("disk files require read/write access");
    }
    const uint64_t max_offset = uint64_t(1) << 63;
    if (data_offset % 4 || scratch_offset % 4 ||
        data_offset > max_offset - 1 - 4 * uint64_t(n) ||
        scratch_offset > max_offset - 1 - 4 * uint64_t(n) ||
        uint64_t(data_info.st_size) < data_offset + 4 * uint64_t(n) ||
        uint64_t(scratch_info.st_size) < scratch_offset + 4 * uint64_t(n))
      throw std::runtime_error("disk file offsets/length");
    std::fill(stats, stats + 12, 0);
    auto begin = std::chrono::steady_clock::now();
    tile = std::min(tile, n);
    std::vector<uint32_t> a(tile), b(tile);
    stats[9] = 4 * (a.capacity() + b.capacity()) + 65536;
    if (a.capacity() > tile || b.capacity() > tile)
      throw std::runtime_error("disk private buffer bound");
    auto ns = [](auto start) {
      return uint64_t(std::chrono::duration_cast<std::chrono::nanoseconds>(
                          std::chrono::steady_clock::now() - start)
                          .count());
    };
    auto read = [&](int fd, uint64_t offset, uint32_t *dst, size_t count) {
      auto io = liftrs::io::read_words(fd, offset, dst, count, r.f.q);
      stats[3] += io.ns;
      stats[5] += io.bytes;
      stats[7] += io.calls;
    };
    auto write = [&](int fd, uint64_t offset, const uint32_t *src, size_t count) {
      auto io = liftrs::io::write_words(fd, offset, src, count);
      stats[4] += io.ns;
      stats[6] += io.bytes;
      stats[8] += io.calls;
    };
    auto measured = [&](unsigned phase, auto fn) {
      auto start = std::chrono::steady_clock::now();
      fn();
      stats[phase] += ns(start);
    };
    auto convert = [&] {
      for (unsigned step = 0; step < dim - split; step++) {
        unsigned k = inverse ? step : dim - split - 1 - step;
        const auto &terms = rows->taylor_terms[k];
        if (terms.empty())
          continue;
        size_t half = size_t(1) << (split + k), minimum = half, maximum = 0;
        for (const auto &t : terms) {
          if (!t.degree || (t.degree & (t.degree - 1)) || t.degree > half / 2)
            throw std::runtime_error("disk Taylor dependency");
          minimum = std::min(minimum, size_t(t.degree));
          maximum = std::max(maximum, size_t(t.degree));
        }
        auto pass = [&](bool quotient) {
          const size_t span = quotient ? maximum : half - minimum;
          liftrs::execution::WorkProgress progress(
              "disk-taylor-tiles", (n / (2 * half)) * ((span + tile - 1) / tile));
          for (size_t block = 0; block < n; block += 2 * half) {
            size_t first = quotient ? 0 : minimum, last = quotient ? maximum : half;
            for (size_t j = first; j < last;) {
              size_t count = std::min(tile, last - j),
                     target = block + (quotient ? half : 0) + j;
              read(data_fd, data_offset + 4 * target, a.data(), count);
              for (const auto &t : terms) {
                size_t d = t.degree, lo = quotient ? j : std::max(j, d),
                       hi = quotient ? std::min(j + count, d) : j + count;
                if (lo >= hi)
                  continue;
                size_t source = block + (quotient ? 2 * half : half) + lo - d;
                read(data_fd, data_offset + 4 * source, b.data(), hi - lo);
                parallel_jobs((hi - lo + 8191) / 8192, workers, [&](size_t job) {
                  for (size_t i = job * 8192; i < std::min(hi - lo, (job + 1) * 8192);
                       i++)
                    a[lo - j + i] ^= t.product.mul(r.f, b[i]);
                });
              }
              write(data_fd, data_offset + 4 * target, a.data(), count);
              j += count;
              progress.advance();
            }
          }
        };
        if (inverse) {
          pass(false);
          pass(true);
        } else {
          pass(true);
          pass(false);
        }
      }
    };
    auto transpose_file = [&](int src, uint64_t src_offset, int dst,
                              uint64_t dst_offset, size_t height, size_t width) {
      size_t block_rows = 1;
      while (block_rows * 2 <= height && (block_rows * 2) <= tile / (block_rows * 2))
        block_rows *= 2;
      size_t block_cols = std::min(width, tile / block_rows);
      liftrs::execution::WorkProgress progress(
          "disk-transpose-blocks", ((height + block_rows - 1) / block_rows) *
                                       ((width + block_cols - 1) / block_cols));
      for (size_t row = 0; row < height; row += block_rows)
        for (size_t col = 0; col < width; col += block_cols) {
          size_t nr = std::min(block_rows, height - row),
                 nc = std::min(block_cols, width - col);
          for (size_t i = 0; i < nr; i++)
            read(src, src_offset + 4 * ((row + i) * width + col), a.data() + i * nc,
                 nc);
          transpose(a.data(), b.data(), nr, nc, workers);
          for (size_t j = 0; j < nc; j++)
            write(dst, dst_offset + 4 * ((col + j) * height + row), b.data() + j * nr,
                  nr);
          progress.advance();
        }
      stats[11]++;
    };
    auto vectors = [&](int fd, uint64_t offset, size_t length, size_t total,
                       const liftrs::Transform &plan, unsigned level, bool affine) {
      size_t batch = tile / length;
      liftrs::execution::WorkProgress progress("disk-vectors", total);
      for (size_t first = 0; first < total;) {
        size_t count = std::min(batch, total - first), words = count * length;
        read(fd, offset + 4 * first * length, a.data(), words);
        parallel_jobs(count, workers, [&](size_t j) {
          plan.recurse(a.data() + j * length, length, level,
                       affine ? uint32_t((first + j) * length) : 0, inverse != 0);
        });
        write(fd, offset + 4 * first * length, a.data(), words);
        first += count;
        progress.advance(count);
      }
    };
    auto columns_phase = [&] {
      transpose_file(data_fd, data_offset, scratch_fd, scratch_offset, n2, n1);
      vectors(scratch_fd, scratch_offset, n2, n1, c, dim - split, false);
      transpose_file(scratch_fd, scratch_offset, data_fd, data_offset, n1, n2);
    };
    auto rows_phase = [&] { vectors(data_fd, data_offset, n1, n2, r, split, true); };
    if (!inverse) {
      measured(0, convert);
      measured(1, columns_phase);
      measured(2, rows_phase);
    } else {
      measured(2, rows_phase);
      measured(1, columns_phase);
      measured(0, convert);
    }
    stats[10] = ns(begin);
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "disk transform failed");
    return 1;
  }
}
