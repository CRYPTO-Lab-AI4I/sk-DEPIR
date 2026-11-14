#pragma once
// Shared32/64-bit schedules. Shape validation precedes all worker dispatch;
// workers borrow disjoint buffers and immutable plan metadata, never Go objects.
#include "progress.hpp"
#include "transform.hpp"
#include <atomic>
#include <chrono>
#include <thread>
namespace liftrs {
namespace execution {
template <class F> void parallel_jobs(size_t jobs, unsigned workers, F fn) {
  WorkProgress progress("parallel-blocks", jobs);
  const unsigned count = std::min<size_t>(workers, jobs);
  if (count <= 1) {
    for (size_t j = 0; j < jobs; j++) {
      fn(j);
      progress.advance();
    }
    return;
  }
  std::atomic<size_t> next{0};
  std::vector<std::thread> threads;
  threads.reserve(count - 1);
  auto run = [&] {
    for (;;) {
      auto j = next.fetch_add(1, std::memory_order_relaxed);
      if (j >= jobs)
        break;
      fn(j);
      progress.advance();
    }
  };
  try {
    for (unsigned k = 1; k < count; k++)
      threads.emplace_back(run);
  } catch (...) {
    for (auto &t : threads)
      t.join();
    throw;
  }
  run();
  for (auto &t : threads)
    t.join();
}
// h-adic conversion: coarse independent blocks use the original recurrence.
// When high levels expose too few blocks, linearized lower degrees d<=half/2
// allow two gathers. Quotient q[j] reads only untouched q[half+j-d] in its
// upper half; remainder then reads the completed quotient. Each destination has
// one writer; joining between passes is essential. Inverse reverses the passes.
template <class Word, class Plan>
void taylor(Word *data, size_t n, const Plan &plan, unsigned split, unsigned v,
            unsigned workers, bool inverse) {
  const auto &h = plan.transform;
  for (unsigned stage = 0; stage < v; stage++) {
    unsigned k = inverse ? stage : v - 1 - stage;
    size_t size = (size_t(1) << split) << (k + 1), half = size / 2, blocks = n / size;
    const auto &terms = plan.taylor_terms[k];
    if (terms.empty())
      continue;
    if (workers > 1 && n >= 32768 && blocks < workers) {
      // Terms are ordered powers of two below the monic degree. Only the first
      // max_degree quotient entries can change; remaining entries stay intact.
      const size_t max_degree = terms.back().degree;
      auto gather = [&](bool quotient) {
        const size_t span = quotient ? max_degree : half, total = blocks * span,
                     chunk = 8192;
        parallel_jobs((total + chunk - 1) / chunk, workers, [&](size_t job) {
          size_t begin = job * chunk, end = std::min(total, begin + chunk);
          while (begin < end) {
            size_t block = begin / span, j = begin % span,
                   count = std::min(end - begin, span - j);
            auto *x = data + block * size;
            for (size_t stop = j + count; j < stop; j++) {
              Word value = x[(quotient ? half : 0) + j];
              for (const auto &t : terms) {
                if (quotient) {
                  if (t.degree > j)
                    value ^= t.product.mul(h.f, x[2 * half + j - t.degree]);
                } else if (t.degree <= j)
                  value ^= t.product.mul(h.f, x[half + j - t.degree]);
              }
              x[(quotient ? half : 0) + j] = value;
            }
            begin += count;
          }
        });
      };
      if (!inverse) {
        gather(true);
        gather(false);
      } else {
        gather(false);
        gather(true);
      }
    } else {
      parallel_jobs(blocks, workers, [&](size_t b) {
        auto *x = data + b * size;
        if (!inverse) {
          for (size_t i = size; i-- > half;) {
            auto z = x[i];
            for (const auto &t : terms)
              x[i - half + t.degree] ^= t.product.mul(h.f, z);
          }
        } else {
          for (size_t i = 0; i < half; i++) {
            auto z = x[half + i];
            for (const auto &t : terms)
              x[i + t.degree] ^= t.product.mul(h.f, z);
          }
        }
      });
    }
  }
}
template <class Word>
void transpose(const Word *src, Word *dst, size_t rows, size_t cols, unsigned workers) {
  constexpr size_t block = 64;
  parallel_jobs((rows + block - 1) / block, workers, [&](size_t group) {
    size_t end = std::min(rows, (group + 1) * block);
    for (size_t c = 0; c < cols; c += block)
      for (size_t r = group * block; r < end; r++)
        for (size_t j = c; j < std::min(cols, c + block); j++)
          dst[j * rows + r] = src[r * cols + j];
  });
}
template <class Word, class Plan>
void matrix(bool bounded, size_t scratch_count, const Plan *rows, const Plan *columns,
            Word *data, Word *scratch, size_t count, unsigned split, unsigned workers,
            int inverse, uint64_t *stage_ns) {
  if (!rows || !columns || !data || !scratch || data == scratch || !workers ||
      workers > 64 || (inverse != 0 && inverse != 1))
    throw std::runtime_error("invalid matrix buffers/workers");
  const auto &r = rows->transform;
  const auto &c = columns->transform;
  unsigned dim = 0;
  for (auto n = count; n > 1; n >>= 1)
    dim++;
  if (!count || (count & (count - 1)) || !split || split >= dim || count > r.f.q ||
      r.terms.size() != split + 1 || rows->taylor_terms.size() != dim - split ||
      c.terms.size() != dim - split || r.f.q != c.f.q ||
      r.f.polynomial != c.f.polynomial || r.f.delta != c.f.delta)
    throw std::runtime_error("invalid matrix shape/fields");
  const size_t n1 = size_t(1) << split, n2 = count / n1;
  auto do_rows = [&] {
    parallel_jobs(n2, workers, [&](size_t j) {
      r.recurse(data + j * n1, n1, split, Word(j * n1), inverse != 0);
    });
  };
  const size_t width = std::min(size_t(64), n1), groups = n1 / width;
  const size_t active = std::min(size_t(workers), groups), tile = width * n2;
  if (bounded && scratch_count < active * tile)
    throw std::runtime_error("short bounded column workspace");
  auto do_columns = [&] {
    if (bounded) {
      // Each worker owns one bounded column batch. Gather/scatter use complete
      // cache-line-sized rows; column transforms remain contiguous in the tile.
      WorkProgress batches("bounded-column-batches", groups);
      parallel_jobs(active, unsigned(active), [&](size_t worker) {
        auto *temp = scratch + worker * tile;
        for (size_t col = worker * width; col < n1; col += active * width) {
          for (size_t row = 0; row < n2; row++)
            for (size_t j = 0; j < width; j++)
              temp[j * n2 + row] = data[row * n1 + col + j];
          for (size_t j = 0; j < width; j++)
            c.recurse(temp + j * n2, n2, dim - split, 0, inverse != 0);
          for (size_t row = 0; row < n2; row++)
            for (size_t j = 0; j < width; j++)
              data[row * n1 + col + j] = temp[j * n2 + row];
          batches.advance();
        }
      });
      return;
    }
    transpose(data, scratch, n2, n1, workers);
    parallel_jobs(n1, workers, [&](size_t i) {
      c.recurse(scratch + i * n2, n2, dim - split, 0, inverse != 0);
    });
    transpose(scratch, data, n1, n2, workers);
  };
  auto measured = [&](unsigned stage, auto fn) {
    auto start = std::chrono::steady_clock::now();
    fn();
    if (stage_ns)
      stage_ns[stage] = std::chrono::duration_cast<std::chrono::nanoseconds>(
                            std::chrono::steady_clock::now() - start)
                            .count();
  };
  auto convert = [&] {
    taylor(data, count, *rows, split, dim - split, workers, inverse != 0);
  };
  if (!inverse) {
    measured(0, convert);
    measured(1, do_columns);
    measured(2, do_rows);
  } else {
    measured(2, do_rows);
    measured(1, do_columns);
    measured(0, convert);
  }
}
} // namespace execution
} // namespace liftrs
