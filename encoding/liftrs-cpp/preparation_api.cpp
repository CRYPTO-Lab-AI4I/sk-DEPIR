//go:build liftrs_cpp && cgo

#include "preparation_api.h"
#include "preparation.hpp"
#include <variant>

struct liftrs_preparation {
  using Narrow = liftrs::preparation::Executor<uint32_t>;
  using Wide = liftrs::preparation::Executor<uint64_t>;
  std::variant<std::unique_ptr<Narrow>, std::unique_ptr<Wide>> state;
};
namespace {
template <class Function> int api(char *error, size_t capacity, Function function) {
  try {
    function();
    return 0;
  } catch (const std::exception &e) {
    if (error && capacity)
      std::snprintf(error, capacity, "%s", e.what());
    return 1;
  }
}
template <class Function> void visit(liftrs_preparation *p, Function function) {
  if (!p)
    throw std::runtime_error("missing preparation handle");
  std::visit(
      [&](auto &state) {
        if (!state)
          throw std::runtime_error("missing preparation state");
        function(*state);
      },
      p->state);
}
} // namespace
extern "C" int liftrs_preparation_create(unsigned s, uint32_t polynomial,
                                         uint32_t delta, uint64_t omega,
                                         unsigned workers, size_t maximum, uint64_t cap,
                                         liftrs_preparation **out, char *error,
                                         size_t capacity) {
  if (out)
    *out = nullptr;
  return api(error, capacity, [&] {
    if (!out)
      throw std::runtime_error("missing preparation output handle");
    auto p = std::make_unique<liftrs_preparation>();
    if (s <= 16)
      p->state = std::make_unique<liftrs_preparation::Narrow>(
          s, polynomial, delta, omega, workers, maximum, cap);
    else
      p->state = std::make_unique<liftrs_preparation::Wide>(s, polynomial, delta, omega,
                                                            workers, maximum, cap);
    *out = p.release();
  });
}
extern "C" int liftrs_preparation_good_count(liftrs_preparation *p, unsigned t,
                                             uint64_t *count, char *error,
                                             size_t capacity) {
  return api(error, capacity, [&] {
    if (!count)
      throw std::runtime_error("missing good count");
    visit(p, [&](auto &state) { *count = state.count_good(t); });
  });
}
extern "C" int liftrs_preparation_good_fill(liftrs_preparation *p, int64_t *output,
                                            size_t count, char *error,
                                            size_t capacity) {
  return api(error, capacity,
             [&] { visit(p, [&](auto &state) { state.fill_good(output, count); }); });
}
extern "C" int liftrs_preparation_good_range(unsigned s, unsigned t, uint64_t first,
                                             unsigned char *output, size_t count,
                                             unsigned workers, char *error,
                                             size_t capacity) {
  return api(error, capacity, [&] {
    if (s < 2 || s > 20 || !t || t >= (uint64_t(1) << s) - 1 || !output || !count ||
        !workers || workers > 64 || first >= (uint64_t(1) << (2 * s)) ||
        count > (uint64_t(1) << (2 * s)) - first)
      throw std::runtime_error("good range shape");
    constexpr size_t chunk = 4096;
    liftrs::execution::parallel_jobs(
        (count + chunk - 1) / chunk, workers, [&](size_t job) {
          for (size_t j = job * chunk; j < std::min(count, (job + 1) * chunk); j++)
            output[j] = liftrs::preparation::good_pair(first + j, s, t);
        });
  });
}
extern "C" int liftrs_preparation_apply(liftrs_preparation *p, void *data, size_t count,
                                        unsigned width, int inverse, char *error,
                                        size_t capacity) {
  return api(error, capacity, [&] {
    if (inverse != 0 && inverse != 1)
      throw std::runtime_error("preparation direction");
    visit(p, [&](auto &state) {
      using Word = typename std::decay_t<decltype(state)>::Plan::word_type;
      if (width != sizeof(Word))
        throw std::runtime_error("preparation word width");
      state.apply(static_cast<Word *>(data), count, inverse != 0);
    });
  });
}
extern "C" int liftrs_preparation_grid(liftrs_preparation *p, void *data, size_t count,
                                       unsigned width, char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(p, [&](auto &state) {
      using Word = typename std::decay_t<decltype(state)>::Plan::word_type;
      if (width != sizeof(Word))
        throw std::runtime_error("preparation word width");
      state.base_grid(static_cast<Word *>(data), count);
    });
  });
}
extern "C" int liftrs_preparation_product(liftrs_preparation *p, const uint32_t *a,
                                          size_t ac, const uint32_t *b, size_t bc,
                                          uint32_t *output, size_t limit, size_t n,
                                          char *error, size_t capacity) {
  return api(error, capacity, [&] {
    visit(p, [&](auto &state) { state.product(a, ac, b, bc, output, limit, n); });
  });
}
extern "C" int liftrs_preparation_roots(liftrs_preparation *p, const int64_t *input,
                                        size_t count, uint32_t *output, char *error,
                                        size_t capacity) {
  return api(error, capacity, [&] {
    visit(p, [&](auto &state) { state.root_factors(input, count, output); });
  });
}
extern "C" int liftrs_preparation_spectrum(liftrs_preparation *p, const uint32_t *g,
                                           size_t count, void *output, size_t n,
                                           unsigned width, char *error,
                                           size_t capacity) {
  return api(error, capacity, [&] {
    visit(p, [&](auto &state) {
      using Word = typename std::decay_t<decltype(state)>::Plan::word_type;
      if (width != sizeof(Word))
        throw std::runtime_error("preparation word width");
      state.spectrum(g, count, static_cast<Word *>(output), n);
    });
  });
}
extern "C" void liftrs_preparation_release_products(liftrs_preparation *p) {
  if (p)
    std::visit(
        [](auto &state) {
          if (state)
            state->release_products();
        },
        p->state);
}
extern "C" uint64_t liftrs_preparation_bytes(const liftrs_preparation *p) {
  return p ? std::visit(
                 [](const auto &state) -> uint64_t {
                   return state ? state->bytes() : 0;
                 },
                 p->state)
           : 0;
}
extern "C" void liftrs_preparation_destroy(liftrs_preparation *p) {
  delete p;
}
