#pragma once
#include <cstdint>
// Exact public-code membership recurrence shared by CPU and CUDA. The Go
// GoodMonomial implementation remains the independent paper criterion oracle.
#if defined(__CUDACC__)
#include <cuda_runtime.h>
#define LIFTRS_MONOMIAL_HD __host__ __device__
#else
#define LIFTRS_MONOMIAL_HD
#endif
namespace liftrs::preparation {
LIFTRS_MONOMIAL_HD inline bool good_weights(const unsigned *weights, unsigned s,
                                            unsigned t) {
  // Lemma4.8: a POSITIVE reachable sum j*(2^s-1) makes the monomial bad.
  // Carry intervals implement that criterion without a q-sized DP per monomial.
  for (unsigned j = 1; j <= 2 * t; j++) {
    unsigned lo = j, hi = j;
    for (unsigned k = 0; k < s; k++) {
      lo = (lo + 1) / 2;
      hi = (hi + t * weights[k]) / 2;
      if (lo > hi)
        break;
    }
    if (lo <= j && j <= hi)
      return false;
  }
  return true;
}
// A ternary digit is the number of set bits at a position across x/y.
// Digit1 has two assignments (10 or01); digits0/2 each have one. Thus each
// class represents2^ones exact monomials, without a q^2 membership bitmap.
LIFTRS_MONOMIAL_HD inline bool good_weight_class(uint64_t index, unsigned s, unsigned t,
                                                 unsigned *ones = nullptr) {
  unsigned weights[20], single = 0;
  for (unsigned k = 0; k < s; k++) {
    weights[k] = unsigned(index % 3);
    index /= 3;
    single += weights[k] == 1;
  }
  if (ones)
    *ones = single;
  return good_weights(weights, s, t);
}
} // namespace liftrs::preparation
#undef LIFTRS_MONOMIAL_HD
