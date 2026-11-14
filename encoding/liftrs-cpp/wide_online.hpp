#pragma once
#include "output.hpp"
// Internal wide C ABI implementation, included once by wide_api.cpp.
struct liftrs_wide_online {
  std::unique_ptr<liftrs_wide_matrix> full, prefix;
  liftrs::Field64 field;
  liftrs::Product64 step;
  uint64_t omega, k, n, l;
  size_t owned_bytes;
};
namespace {
struct wide_range {
  uintptr_t begin, end;
};
wide_range range(const void *p, size_t bytes) {
  auto begin = reinterpret_cast<uintptr_t>(p);
  if (!p || begin > std::numeric_limits<uintptr_t>::max() - bytes)
    throw std::runtime_error("invalid wide memory range");
  return {begin, begin + bytes};
}
bool overlaps(wide_range a, wide_range b) {
  return a.begin < b.end && b.begin < a.end;
}
} // namespace
extern "C" int liftrs_wide_online_create(unsigned s, uint32_t polynomial,
                                         uint32_t delta, uint64_t omega, uint64_t k,
                                         liftrs_wide_online **out, char *error,
                                         size_t capacity) {
  if (out)
    *out = nullptr;
  try {
    if (!out)
      throw std::runtime_error("missing wide encoder handle");
    auto e = std::make_unique<liftrs_wide_online>();
    field(e->field, s, polynomial, delta);
    e->n = e->field.q;
    e->k = k;
    e->omega = omega;
    if (!k || k > e->n / 2 || !liftrs::is_primitive(e->field, omega))
      throw std::runtime_error("invalid wide online shape/generator");
    e->l = 1;
    unsigned prefix_dim = 0;
    while (e->l < 2 * k - 1) {
      e->l *= 2;
      prefix_dim++;
    }
    liftrs_wide_matrix *matrix = nullptr;
    if (liftrs_wide_matrix_create(s, polynomial, delta, 2 * s, s, &matrix, error,
                                  capacity))
      throw std::runtime_error(error);
    e->full.reset(matrix);
    if (e->l > 1 && e->l != e->n) {
      matrix = nullptr;
      if (liftrs_wide_matrix_create(s, polynomial, delta, prefix_dim,
                                    std::max(1u, prefix_dim / 2), &matrix, error,
                                    capacity))
        throw std::runtime_error(error);
      e->prefix.reset(matrix);
    }
    e->step = liftrs::constant_product(e->field, omega);
    e->owned_bytes = sizeof(*e) + e->step.rows.capacity() * 8 +
                     liftrs_wide_matrix_bytes(e->full.get()) +
                     liftrs_wide_matrix_bytes(e->prefix.get());
    *out = e.release();
    return 0;
  } catch (const std::exception &ex) {
    fail(error, capacity, ex.what());
    return 1;
  }
}
extern "C" int liftrs_wide_online_encode(
    const liftrs_wide_online *e, unsigned workers, const uint32_t *message,
    const uint32_t *p, const uint32_t *spectrum32, const uint64_t *spectrum64,
    uint64_t *arena, size_t arena_count, uint32_t *h, size_t h_count, uint64_t *scratch,
    size_t scratch_count, uint64_t *stats, char *error, size_t capacity) {
  try {
    if (!e || !stats || !workers || workers > 64 || !message || !p || !arena || !h ||
        !scratch || arena_count != std::max(2 * e->l, e->n) || h_count != e->k ||
        scratch_count != e->n || (bool(spectrum32) == bool(spectrum64)))
      throw std::runtime_error("invalid wide online buffers/width");
    const auto begin = std::chrono::steady_clock::now();
    std::fill(stats, stats + 7, 0);
    const wide_range writes[] = {range(arena, arena_count * 8), range(h, h_count * 4),
                                 range(scratch, scratch_count * 8)};
    const wide_range reads[] = {
        range(message, e->k * 4), range(p, (e->k + 1) * 4),
        range(spectrum32 ? static_cast<const void *>(spectrum32) : spectrum64,
              e->n * (spectrum32 ? 4 : 8))};
    for (size_t i = 0; i < 3; i++) {
      for (size_t j = i + 1; j < 3; j++)
        if (overlaps(writes[i], writes[j]))
          throw std::runtime_error("wide online workspace overlap");
      for (auto r : reads)
        if (overlaps(writes[i], r))
          throw std::runtime_error("wide online input/workspace overlap");
    }
    auto measured = [&](unsigned index, auto fn) {
      auto start = std::chrono::steady_clock::now();
      fn();
      stats[index] += elapsed(start);
    };
    const uint64_t base_q = uint64_t(1) << e->field.s;
    auto prepare = [&] {
      measured(4, [&] {
        if (p[0] != 1)
          throw std::runtime_error("wide P normalization");
        for (size_t i = 0; i < e->k; i++)
          if (message[i] >= base_q || p[i] >= base_q)
            throw std::runtime_error("wide online base value range");
        if (p[e->k] >= base_q)
          throw std::runtime_error("wide P end range");
        for (size_t i = 0; i < e->n; i++)
          if ((spectrum32 ? uint64_t(spectrum32[i]) : spectrum64[i]) >= e->n)
            throw std::runtime_error("wide spectrum range");
        std::fill(arena, arena + 2 * e->l, 0);
        for (size_t i = 0; i < e->k; i++) {
          arena[i] = p[i];
          arena[e->l + i] = message[i];
        }
      });
    };
    auto fft = [&](bool prefix, uint64_t *data, bool inverse) {
      if (prefix && e->l == 1)
        return;
      const auto *m = prefix && e->prefix ? e->prefix.get() : e->full.get();
      uint64_t stages[3] = {};
      liftrs::execution::matrix(false, prefix ? e->l : e->n, m->rows.get(),
                                m->columns.get(), data, scratch, prefix ? e->l : e->n,
                                m->split, workers, inverse, stages);
      for (unsigned j = 0; j < 3; j++)
        stats[1 + j] += stages[j];
    };
    const size_t chunk = 4096;
    auto prefix_product = [&] {
      measured(5, [&] {
        liftrs::execution::pointwise(e->field, arena, e->l, workers, chunk,
                                     [&](size_t i) { return arena[e->l + i]; });
      });
    };
    auto promote = [&] {
      measured(4, [&] {
        for (size_t i = 0; i < e->k; ++i) {
          if (arena[i] >= base_q)
            throw std::runtime_error("wide prefix escaped base field");
          h[i] = uint32_t(arena[i]);
        }
        std::fill(arena, arena + e->n, 0);
        for (size_t i = 0; i < e->k; i++)
          arena[i] = h[i];
      });
    };
    auto spectrum_product = [&] {
      measured(5, [&] {
        liftrs::execution::pointwise(
            e->field, arena, e->n, workers, chunk, [&](size_t i) {
              return spectrum32 ? uint64_t(spectrum32[i]) : spectrum64[i];
            });
      });
    };
    auto assemble = [&] {
      measured(6, [&] {
        // Scratch is dead. Character access may alias its uint64 representation.
        // Use relaxed atomic byte stores: two output words can share one original
        // uint64 memory location even though primitive-generator scatter is bijective.
        // No typed uint64 access occurs until a later borrowed call reuses scratch.
        auto *output = reinterpret_cast<unsigned char *>(scratch);
        const size_t jobs = (e->n - 1 + chunk - 1) / chunk;
        std::vector<uint32_t> parity(jobs);
        std::atomic<bool> bad{false};
        liftrs::execution::scatter(
            e->field, e->omega, e->step, 0, e->n - 1, workers, chunk,
            [&](size_t i) { return arena[i]; },
            [&](uint64_t point, uint32_t value) {
              liftrs::output::store_native_u32(output, size_t(point), value);
            },
            [&](size_t job, uint32_t sum) { parity[job] = sum; }, bad);
        if (bad.load(std::memory_order_relaxed))
          throw std::runtime_error("wide product escaped base field");
        uint32_t sum = 0;
        for (auto x : parity)
          sum ^= x;
        std::memcpy(output, &sum, 4);
      });
    };
    liftrs::execution::encode_cyclic(
        prepare,
        [&](liftrs::execution::Input input, bool inverse) {
          fft(input != liftrs::execution::Input::full,
              input == liftrs::execution::Input::prefix_message ? arena + e->l : arena,
              inverse);
        },
        prefix_product, promote, spectrum_product, assemble);
    stats[0] = elapsed(begin);
    return 0;
  } catch (const std::exception &ex) {
    fail(error, capacity, ex.what());
    return 1;
  }
}
extern "C" size_t liftrs_wide_online_bytes(const liftrs_wide_online *e) {
  return e ? e->owned_bytes : 0;
}
extern "C" void liftrs_wide_online_destroy(liftrs_wide_online *e) {
  delete e;
}
