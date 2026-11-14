#pragma once
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
} // namespace liftrs::preparation
#undef LIFTRS_MONOMIAL_HD
