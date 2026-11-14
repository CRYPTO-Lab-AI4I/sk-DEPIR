//go:build ignore

// Include the implementation only to test private Taylor scheduling, without
// adding a test-only ABI. The oracle uses generic field products, not maps.
#include "c_api.cpp"
#include <iostream>
#include <random>
void reference(std::vector<uint32_t> &data, const liftrs_plan &plan, unsigned split,
               unsigned v, bool inverse) {
  for (unsigned stage = 0; stage < v; stage++) {
    unsigned k = inverse ? stage : v - 1 - stage;
    size_t size = size_t(1) << (split + k + 1), half = size / 2;
    for (size_t block = 0; block < data.size(); block += size) {
      auto *x = data.data() + block;
      if (!inverse) {
        for (size_t i = size; i-- > half;)
          for (const auto &t : plan.taylor_terms[k])
            x[i - half + t.degree] ^= plan.transform.f.mul(t.product.coefficient, x[i]);
      } else {
        for (size_t i = 0; i < half; i++)
          for (const auto &t : plan.taylor_terms[k])
            x[i + t.degree] ^= plan.transform.f.mul(t.product.coefficient, x[half + i]);
      }
    }
  }
}
int main() {
  std::mt19937 random(2301);
  size_t checks = 0;
  for (bool extension : {false, true})
    for (unsigned dim : {14u, 15u, 16u})
      for (unsigned split : {1u, 5u, 10u}) {
        liftrs_plan plan;
        auto &f = plan.transform.f;
        f.s = 4;
        f.polynomial = 0x13;
        f.delta = 8;
        f.q = extension ? 256 : 16;
        unsigned v = dim - split;
        plan.taylor_terms.resize(v);
        for (unsigned k = 0; k < v; k++)
          for (unsigned j = 0; j < split; j++) {
            uint32_t coefficient = (j + k) % 3 == 0   ? 0
                                   : (j + k) % 3 == 1 ? 1
                                                      : random() % f.q;
            plan.taylor_terms[k].push_back(
                {uint32_t(1) << (j + k), liftrs::constant_product(f, coefficient)});
          }
        std::vector<uint32_t> input(size_t(1) << dim);
        for (auto &x : input)
          x = random() % f.q;
        auto want = input;
        reference(want, plan, split, v, false);
        auto inverse_want = input;
        reference(inverse_want, plan, split, v, true);
        for (unsigned workers : {1u, 2u, 8u, 32u}) {
          auto got = input;
          liftrs::execution::taylor(got.data(), got.size(), plan, split, v, workers,
                                    false);
          if (got != want)
            throw std::runtime_error("Taylor gather/reference mismatch");
          liftrs::execution::taylor(got.data(), got.size(), plan, split, v, workers,
                                    true);
          if (got != input)
            throw std::runtime_error("Taylor inverse mismatch");
          // Independent inverse input avoids relying solely on a round trip.
          got = input;
          liftrs::execution::taylor(got.data(), got.size(), plan, split, v, workers,
                                    true);
          if (got != inverse_want)
            throw std::runtime_error("Taylor inverse/reference mismatch");
          checks++;
        }
      }
  std::cout << checks << " Taylor reference/forward/inverse checks passed\n";
}
