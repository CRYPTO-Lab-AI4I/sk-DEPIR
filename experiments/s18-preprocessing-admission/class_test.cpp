// Bounded host self-check for class multiplicity, not the independent Go oracle.
#include "../../encoding/liftrs-cpp/monomial.hpp"
#include <cstdint>
#include <iostream>
#include <stdexcept>
int main() {
  for (unsigned s = 2; s <= 8; s++) {
    uint64_t classes = 1;
    for (unsigned j = 0; j < s; j++)
      classes *= 3;
    for (unsigned t : {1u, 3u}) {
      if (t >= (1u << s) - 1)
        continue;
      uint64_t dimension = 0, universe = 0;
      for (uint64_t i = 0; i < classes; i++) {
        unsigned ones = 0;
        bool good = liftrs::preparation::good_weight_class(i, s, t, &ones);
        universe += 1ull << ones;
        if (good)
          dimension += 1ull << ones;
      }
      uint64_t direct = 0;
      for (uint64_t i = 0; i < (1ull << (2 * s)); i++) {
        unsigned weights[20];
        for (unsigned k = 0; k < s; k++)
          weights[k] = unsigned((i >> k) & 1) + unsigned((i >> (s + k)) & 1);
        direct += liftrs::preparation::good_weights(weights, s, t);
      }
      if (universe != (1ull << (2 * s)) || direct != dimension)
        throw std::runtime_error("class cardinality mismatch");
    }
  }
  std::cout << "class cardinalities and weighted/direct membership match s2..8"
            << std::endl;
}
