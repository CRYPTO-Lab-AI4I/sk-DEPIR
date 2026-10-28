//go:build liftrs_cpp && cgo

#include "c_api.h"
#include "transform.hpp"
#include <chrono>
#include <cstdio>
#include <memory>

struct liftrs_plan {
  liftrs::Transform transform;
  size_t owned_bytes = 0; // Vector capacity payloads + descriptors, not allocator RSS.
};
namespace {
struct Reader {
  const uint32_t* data;
  size_t count, position = 0;
  uint32_t word() {
    if (position == count) throw std::runtime_error("short native plan");
    return data[position++];
  }
};
void fail(char* error, size_t capacity, const char* message) {
  if (error && capacity) std::snprintf(error, capacity, "%s", message);
}
liftrs::Product product(Reader& r, const liftrs::Field& f, unsigned width,
                         size_t& bytes) {
  liftrs::Product p;
  p.coefficient = r.word();
  const auto count = r.word();
  if (p.coefficient >= f.q || (count && count != ((width+3)/4)*16))
    throw std::runtime_error("invalid native multiplication map");
  p.rows.resize(count);
  for (auto& x : p.rows) {
    x = r.word();
    if (x >= f.q) throw std::runtime_error("native table value outside field");
  }
  bytes += p.rows.capacity()*sizeof(uint32_t);
  return p;
}
}
extern "C" int liftrs_plan_create(const uint32_t* words, size_t count,
    liftrs_plan** out, char* error, size_t capacity) {
  if (out) *out = nullptr;
  try {
    if (!out || !words || count > (1u<<20))
      throw std::runtime_error("invalid native plan size");
    Reader r{words,count};
    if (r.word() != 1) throw std::runtime_error("unsupported native plan version");
    auto plan = std::make_unique<liftrs_plan>();
    auto& a = plan->transform;
    a.f.s = r.word(); a.f.polynomial = r.word(); a.f.delta = r.word();
    a.f.q = r.word(); const unsigned levels = r.word();
    if (a.f.s < 2 || a.f.s > 13 ||
        (levels != a.f.s && levels != 2*a.f.s) || a.f.q != (1u<<levels) ||
        (a.f.polynomial >> a.f.s) != 1 || !(a.f.polynomial&1) ||
        a.f.delta >= (1u<<a.f.s))
      throw std::runtime_error("invalid native field description");
    a.terms.resize(levels); a.values.resize(levels);
    a.delta.resize(levels); a.inverse.resize(levels);
    a.delta_products.resize(levels); a.inverse_products.resize(levels);
    auto& bytes = plan->owned_bytes;
    bytes = sizeof(liftrs_plan) + a.terms.capacity()*sizeof(a.terms[0]) +
      a.values.capacity()*sizeof(a.values[0]) +
      (a.delta.capacity()+a.inverse.capacity())*sizeof(uint32_t) +
      (a.delta_products.capacity()+a.inverse_products.capacity())*sizeof(liftrs::Product);
    for (unsigned k=0;k<levels;k++) {
      const unsigned terms = r.word();
      if (!terms || terms > k+2) throw std::runtime_error("invalid native terms");
      a.terms[k].reserve(terms);
      uint32_t previous = 0;
      for (unsigned j=0;j<terms;j++) {
        liftrs::Term t; t.degree = r.word();
        if (t.degree > (1u<<k) || (j && t.degree <= previous))
          throw std::runtime_error("invalid native term degree");
        previous = t.degree;
        t.product = product(r,a.f,levels,bytes);
        a.terms[k].push_back(std::move(t));
      }
      const auto& leading = a.terms[k].back();
      if (leading.degree != (1u<<k) || leading.product.coefficient != 1)
        throw std::runtime_error("nonmonic native transform polynomial");
      bytes += a.terms[k].capacity()*sizeof(liftrs::Term);
      a.values[k].resize(levels);
      for (auto& x : a.values[k]) {
        x = r.word();
        if (x >= a.f.q) throw std::runtime_error("invalid native offset value");
      }
      bytes += a.values[k].capacity()*sizeof(uint32_t);
      a.delta[k] = r.word(); a.inverse[k] = r.word();
      if (!a.delta[k] || !a.inverse[k] || a.delta[k]>=a.f.q || a.inverse[k]>=a.f.q)
        throw std::runtime_error("invalid native butterfly factor");
      a.delta_products[k] = product(r,a.f,levels,bytes);
      a.inverse_products[k] = product(r,a.f,levels,bytes);
    }
    if (r.position != count) throw std::runtime_error("trailing native plan data");
    *out = plan.release(); return 0;
  } catch (const std::exception& e) { fail(error,capacity,e.what()); return 1; }
    catch (...) { fail(error,capacity,"native initialization failed"); return 1; }
}
extern "C" int liftrs_plan_apply(const liftrs_plan* plan, uint32_t* data,
    size_t count, int inverse, uint64_t* kernel_ns, char* error, size_t capacity) {
  try {
    if (!plan || !data || !kernel_ns || (inverse!=0 && inverse!=1))
      throw std::runtime_error("invalid native transform call");
    const auto start = std::chrono::steady_clock::now();
    plan->transform.apply(data,count,inverse!=0);
    *kernel_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(
      std::chrono::steady_clock::now()-start).count();
    return 0;
  } catch (const std::exception& e) { fail(error,capacity,e.what()); return 1; }
    catch (...) { fail(error,capacity,"native transform failed"); return 1; }
}
extern "C" size_t liftrs_plan_owned_bytes(const liftrs_plan* plan) {
  return plan ? plan->owned_bytes : 0;
}
extern "C" void liftrs_plan_destroy(liftrs_plan* plan) { delete plan; }
