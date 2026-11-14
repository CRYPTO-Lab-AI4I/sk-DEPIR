//go:build liftrs_cpp && cgo

#include "c_api.h"
#include "file_io.hpp"
#include "matrix.hpp"
#include "online.hpp"
#include "plan.hpp"
#include <atomic>
#include <cerrno>
#include <chrono>
#include <cstdio>
#include <cstring>
#include <memory>
#include <thread>
#include <time.h>
#include <unistd.h>

namespace {
struct Reader {
  const uint32_t *data;
  size_t count, position = 0;
  uint32_t word() {
    if (position == count)
      throw std::runtime_error("short native plan");
    return data[position++];
  }
};
void fail(char *error, size_t capacity, const char *message) {
  if (error && capacity)
    std::snprintf(error, capacity, "%s", message);
}
liftrs::Product product(Reader &r, const liftrs::Field &f, unsigned width,
                        size_t &bytes) {
  liftrs::Product p;
  p.coefficient = r.word();
  const auto count = r.word();
  if (p.coefficient >= f.q || (count && count != ((width + 3) / 4) * 16))
    throw std::runtime_error("invalid native multiplication map");
  p.rows.resize(count);
  for (auto &x : p.rows) {
    x = r.word();
    if (x >= f.q)
      throw std::runtime_error("native table value outside field");
  }
  bytes += p.rows.capacity() * sizeof(uint32_t);
  return p;
}
} // namespace
extern "C" int liftrs_plan_create(const uint32_t *words, size_t count,
                                  liftrs_plan **out, char *error, size_t capacity) {
  if (out)
    *out = nullptr;
  try {
    if (!out || !words || count > (1u << 20))
      throw std::runtime_error("invalid native plan size");
    Reader r{words, count};
    const auto version = r.word();
    if (version < 1 || version > 4)
      throw std::runtime_error("unsupported native plan version");
    auto plan = std::make_unique<liftrs_plan>();
    auto &a = plan->transform;
    a.f.s = r.word();
    a.f.polynomial = r.word();
    a.f.delta = r.word();
    a.f.q = r.word();
    if (version >= 3)
      a.f.q |= uint64_t(r.word()) << 32;
    const unsigned levels = r.word();
    unsigned width = 0;
    for (auto q = a.f.q; q > 1; q >>= 1)
      width++;
    if (a.f.s < 2 || a.f.s > 16 || width > 32 ||
        (width != a.f.s && width != 2 * a.f.s) || a.f.q != (uint64_t(1) << width) ||
        !levels || levels > width || ((version & 1) && levels != width) ||
        (a.f.polynomial >> a.f.s) != 1 || !(a.f.polynomial & 1) ||
        a.f.delta >= (1u << a.f.s))
      throw std::runtime_error("invalid native field description");
    a.terms.resize(levels);
    a.values.resize(levels);
    a.delta.resize(levels);
    a.inverse.resize(levels);
    a.delta_products.resize(levels);
    a.inverse_products.resize(levels);
    if (!(version & 1)) {
      a.basis.resize(levels);
      for (auto &x : a.basis) {
        x = r.word();
        if (!x || x >= a.f.q)
          throw std::runtime_error("invalid ordered basis");
      }
    }
    auto &bytes = plan->owned_bytes;
    bytes = a.basis.capacity() * sizeof(uint32_t) + sizeof(liftrs_plan) +
            a.terms.capacity() * sizeof(a.terms[0]) +
            a.values.capacity() * sizeof(a.values[0]) +
            (a.delta.capacity() + a.inverse.capacity()) * sizeof(uint32_t) +
            (a.delta_products.capacity() + a.inverse_products.capacity()) *
                sizeof(liftrs::Product);
    for (unsigned k = 0; k < levels; k++) {
      const unsigned terms = r.word();
      if (!terms || terms > k + 2)
        throw std::runtime_error("invalid native terms");
      a.terms[k].reserve(terms);
      uint32_t previous = 0;
      for (unsigned j = 0; j < terms; j++) {
        liftrs::Term t;
        t.degree = r.word();
        if (t.degree > (1u << k) || (j && t.degree <= previous))
          throw std::runtime_error("invalid native term degree");
        previous = t.degree;
        t.product = product(r, a.f, width, bytes);
        a.terms[k].push_back(std::move(t));
      }
      const auto &leading = a.terms[k].back();
      if (leading.degree != (1u << k) || leading.product.coefficient != 1)
        throw std::runtime_error("nonmonic native transform polynomial");
      bytes += a.terms[k].capacity() * sizeof(liftrs::Term);
      a.values[k].resize(width);
      for (auto &x : a.values[k]) {
        x = r.word();
        if (x >= a.f.q)
          throw std::runtime_error("invalid native offset value");
      }
      bytes += a.values[k].capacity() * sizeof(uint32_t);
      a.delta[k] = r.word();
      a.inverse[k] = r.word();
      if (!a.delta[k] || !a.inverse[k] || a.delta[k] >= a.f.q || a.inverse[k] >= a.f.q)
        throw std::runtime_error("invalid native butterfly factor");
      a.delta_products[k] = product(r, a.f, width, bytes);
      a.inverse_products[k] = product(r, a.f, width, bytes);
    }
    if (r.position != count)
      throw std::runtime_error("trailing native plan data");
    *out = plan.release();
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "native initialization failed");
    return 1;
  }
}
extern "C" int liftrs_plan_apply(const liftrs_plan *plan, uint32_t *data, size_t count,
                                 int inverse, uint64_t *kernel_ns, char *error,
                                 size_t capacity) {
  try {
    if (!plan || !data || !kernel_ns || (inverse != 0 && inverse != 1))
      throw std::runtime_error("invalid native transform call");
    const auto start = std::chrono::steady_clock::now();
    plan->transform.apply(data, count, inverse != 0);
    *kernel_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
                     std::chrono::steady_clock::now() - start)
                     .count();
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "native transform failed");
    return 1;
  }
}
extern "C" size_t liftrs_plan_owned_bytes(const liftrs_plan *plan) {
  return plan ? plan->owned_bytes : 0;
}
extern "C" void liftrs_plan_destroy(liftrs_plan *plan) {
  delete plan;
}

extern "C" int liftrs_plan_apply_at(const liftrs_plan *plan, uint32_t *data,
                                    size_t count, uint32_t offset, int inverse,
                                    uint64_t *ns, char *error, size_t capacity) {
  try {
    if (!plan || !data || !ns || (inverse != 0 && inverse != 1))
      throw std::runtime_error("invalid affine call");
    const auto start = std::chrono::steady_clock::now();
    plan->transform.apply(data, count, inverse != 0, offset);
    *ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
              std::chrono::steady_clock::now() - start)
              .count();
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  }
}
namespace {
// Each job owns disjoint output. Exceptions during thread creation join already
// launched workers; mathematical kernels cannot throw once shape is validated.
using liftrs::execution::parallel_jobs;
using liftrs::execution::transpose;
uint32_t frobenius(const liftrs::Field &f, uint32_t c, unsigned k) {
  while (k--)
    c = f.mul(c, c);
  return c;
}

} // namespace
static int matrix_apply(bool bounded, size_t scratch_count, const liftrs_plan *rows,
                        const liftrs_plan *columns, uint32_t *data, uint32_t *scratch,
                        size_t count, unsigned split, unsigned workers, int inverse,
                        uint64_t *stage_ns, char *error, size_t capacity) {
  try {
    liftrs::execution::matrix(bounded, scratch_count, rows, columns, data, scratch,
                              count, split, workers, inverse, stage_ns);
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  } catch (...) {
    fail(error, capacity, "native matrix execution failed");
    return 1;
  }
}

extern "C" int liftrs_matrix_apply(const liftrs_plan *rows, const liftrs_plan *columns,
                                   uint32_t *data, uint32_t *scratch, size_t count,
                                   unsigned split, unsigned workers, int inverse,
                                   uint64_t *stages, char *error, size_t capacity) {
  return matrix_apply(false, count, rows, columns, data, scratch, count, split, workers,
                      inverse, stages, error, capacity);
}

extern "C" int liftrs_plan_prepare_taylor(liftrs_plan *plan, unsigned high_levels,
                                          char *error, size_t capacity) {
  try {
    if (!plan || !high_levels || !plan->taylor_terms.empty())
      throw std::runtime_error("invalid/repeated Taylor preparation");
    const auto &h = plan->transform;
    unsigned split = h.terms.size() - 1, width = 0;
    for (auto q = h.f.q; q > 1; q >>= 1)
      width++;
    if (split + high_levels > width)
      throw std::runtime_error("Taylor domain exceeds field");
    std::vector<std::vector<liftrs::Term>> prepared(high_levels);
    size_t bytes = prepared.capacity() * sizeof(prepared[0]);
    for (unsigned k = 0; k < high_levels; k++) {
      auto &ts = prepared[k];
      const auto &source = h.terms[split];
      ts.reserve(source.size() - 1);
      for (size_t j = 0; j + 1 < source.size(); j++)
        ts.push_back({source[j].degree << k,
                      liftrs::constant_product(
                          h.f, frobenius(h.f, source[j].product.coefficient, k))});
      bytes += ts.capacity() * sizeof(liftrs::Term);
      for (const auto &t : ts)
        bytes += t.product.rows.capacity() * sizeof(uint32_t);
    }
    plan->taylor_terms = std::move(prepared);
    plan->owned_bytes += bytes;
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  }
}

// ABI fragments preserve the established standalone c_api.cpp build command.
#include "clocks.hpp"
#include "disk.hpp"
#include "online32.hpp"
