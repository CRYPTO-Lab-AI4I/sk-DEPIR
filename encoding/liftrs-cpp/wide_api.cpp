//go:build liftrs_cpp && cgo

#include "wide_api.h"
#include "online.hpp"
#include "prepare_plan.hpp"
#include <cstdio>
#include <cstring>
#include <limits>
#include <memory>

struct liftrs_wide_plan : liftrs::TransformPlan<uint64_t> {};
struct liftrs_wide_matrix : liftrs::MatrixPlan<uint64_t, liftrs_wide_plan> {};
namespace {
void fail(char *error, size_t capacity, const char *text) {
  if (error && capacity)
    std::snprintf(error, capacity, "%s", text);
}
void field(liftrs::Field64 &f, unsigned s, uint32_t polynomial, uint32_t delta) {
  liftrs::preparation::set_field(f, s, polynomial, delta);
}
void inputs(const liftrs::Field64 &f, const uint64_t *data, size_t count) {
  for (size_t i = 0; i < count; i++)
    if (data[i] >= f.q)
      throw std::runtime_error("wide input outside field");
}
uint64_t elapsed(std::chrono::steady_clock::time_point start) {
  return std::chrono::duration_cast<std::chrono::nanoseconds>(
             std::chrono::steady_clock::now() - start)
      .count();
}
} // namespace
extern "C" int liftrs_wide_plan_create(unsigned s, uint32_t polynomial, uint32_t delta,
                                       const uint64_t *basis, size_t levels,
                                       liftrs_wide_plan **out, char *error,
                                       size_t capacity) {
  if (out)
    *out = nullptr;
  try {
    if (!out)
      throw std::runtime_error("missing wide handle");
    *out = liftrs::preparation::build_transform<liftrs_wide_plan>(s, polynomial, delta,
                                                                  basis, levels)
               .release();
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  }
}
extern "C" int liftrs_wide_plan_apply(const liftrs_wide_plan *p, uint64_t *data,
                                      size_t count, uint64_t start, int inverse,
                                      uint64_t *ns, char *error, size_t capacity) {
  try {
    if (!p || !data || !ns || (inverse != 0 && inverse != 1) || !count ||
        (count & (count - 1)) || count > (uint64_t(1) << p->transform.terms.size()) ||
        start >= p->transform.f.q)
      throw std::runtime_error("invalid wide transform call");
    auto begin = std::chrono::steady_clock::now();
    inputs(p->transform.f, data, count);
    p->transform.apply(data, count, inverse != 0, start);
    *ns = elapsed(begin);
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  }
}
extern "C" size_t liftrs_wide_plan_bytes(const liftrs_wide_plan *p) {
  return p ? p->owned_bytes : 0;
}
extern "C" void liftrs_wide_plan_destroy(liftrs_wide_plan *p) {
  delete p;
}
extern "C" int liftrs_wide_matrix_create(unsigned s, uint32_t polynomial,
                                         uint32_t delta, unsigned dim, unsigned split,
                                         liftrs_wide_matrix **out, char *error,
                                         size_t capacity) {
  if (out)
    *out = nullptr;
  try {
    if (!out || !split || split >= dim || s < 2 || s > 20 || dim > 2 * s)
      throw std::runtime_error("invalid wide matrix dimensions");
    auto m = liftrs::preparation::build_matrix<liftrs_wide_matrix>(s, polynomial, delta,
                                                                   dim, split);
    *out = m.release();
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  }
}
extern "C" int liftrs_wide_matrix_apply(const liftrs_wide_matrix *m, uint64_t *data,
                                        uint64_t *scratch, size_t count,
                                        unsigned workers, int inverse, uint64_t *ns,
                                        uint64_t *stages, char *error,
                                        size_t capacity) {
  try {
    if (!m || !data || !scratch || !ns || !stages || !workers || workers > 64 ||
        (inverse != 0 && inverse != 1) || count != (uint64_t(1) << m->dim))
      throw std::runtime_error("invalid wide matrix call");
    const auto bytes = count * sizeof(uint64_t), a = reinterpret_cast<uintptr_t>(data),
               b = reinterpret_cast<uintptr_t>(scratch);
    if (a > std::numeric_limits<uintptr_t>::max() - bytes ||
        b > std::numeric_limits<uintptr_t>::max() - bytes ||
        (a < b + bytes && b < a + bytes))
      throw std::runtime_error("wide matrix buffers overlap");
    auto begin = std::chrono::steady_clock::now();
    inputs(m->rows->transform.f, data, count);
    liftrs::execution::matrix(false, count, m->rows.get(), m->columns.get(), data,
                              scratch, count, m->split, workers, inverse, stages);
    *ns = elapsed(begin);
    return 0;
  } catch (const std::exception &e) {
    fail(error, capacity, e.what());
    return 1;
  }
}
extern "C" size_t liftrs_wide_matrix_bytes(const liftrs_wide_matrix *m) {
  return m ? sizeof(*m) + m->rows->owned_bytes + m->columns->owned_bytes : 0;
}
extern "C" void liftrs_wide_matrix_destroy(liftrs_wide_matrix *m) {
  delete m;
}

// Complete wide encoding shares online.hpp with all32-bit storage modes.
#include "wide_online.hpp"
