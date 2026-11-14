#pragma once
#include "transform.hpp"

namespace liftrs {

template <class Word>
Word field_power(const FieldT<Word> &field, Word x, uint64_t exponent) {
  Word result = 1;
  while (exponent) {
    if (exponent & 1)
      result = field.mul(result, x);
    exponent >>= 1;
    if (exponent)
      x = field.mul(x, x);
  }
  return result;
}

// Factor the multiplicative-group order, never the field's domain arrays.
// Division in the loop condition avoids overflow for wider cardinalities.
template <class Word> bool is_primitive(const FieldT<Word> &field, Word omega) {
  const uint64_t order = field.q - 1;
  if (!omega || uint64_t(omega) >= field.q || field_power(field, omega, order) != 1)
    return false;
  auto remaining = order;
  for (uint64_t d = 2; d <= remaining / d; ++d) {
    if (remaining % d)
      continue;
    if (field_power(field, omega, order / d) == 1)
      return false;
    while (remaining % d == 0)
      remaining /= d;
  }
  return remaining <= 1 || field_power(field, omega, order / remaining) != 1;
}

template <class Word>
ProductT<Word> constant_product(const FieldT<Word> &field, Word coefficient) {
  ProductT<Word> product;
  product.coefficient = coefficient;
  if (coefficient > 1) {
    unsigned width = 0;
    for (auto q = field.q; q > 1; q >>= 1)
      ++width;
    product.rows.resize(((width + 3) / 4) * 16);
    for (size_t j = 0; j < product.rows.size(); ++j)
      product.rows[j] =
          field.mul(coefficient, (Word(j % 16) << (4 * (j / 16))) & Word(field.q - 1));
  }
  return product;
}

} // namespace liftrs
