#pragma once
// Sparse field/transform/matrix construction shared by preprocessing widths and
// the wide C ABI. All allocations are metadata-sized; no field grid is built.
#include "arithmetic.hpp"
#include "plan.hpp"

namespace liftrs::preparation {
template <class Word>
void set_field(FieldT<Word> &f, unsigned s, uint32_t polynomial, uint32_t delta,
               bool base = false) {
  if (sizeof(size_t) < 8 || s < 2 || s > 20 || (base ? s : 2 * s) > 8 * sizeof(Word) ||
      (polynomial >> s) != 1 || !(polynomial & 1) || !delta || delta >= (1u << s))
    throw std::runtime_error("invalid wide field");
  f.s = s;
  f.polynomial = polynomial;
  f.delta = delta;
  f.q = uint64_t(1) << (base ? s : 2 * s);
}
template <class Word>
Word evaluate_linear(const FieldT<Word> &f, const std::vector<Word> &p, Word x) {
  Word out = 0, v = x;
  for (auto c : p) {
    out ^= f.mul(c, v);
    v = f.mul(v, v);
  }
  return out;
}
template <class Plan> void account_plan(Plan &p) {
  using Word = typename Plan::word_type;
  const auto &a = p.transform;
  size_t bytes = sizeof(p) + a.basis.capacity() * sizeof(Word) +
                 (a.delta.capacity() + a.inverse.capacity()) * sizeof(Word);
  bytes += a.terms.capacity() * sizeof(a.terms[0]) +
           a.values.capacity() * sizeof(a.values[0]);
  bytes += (a.delta_products.capacity() + a.inverse_products.capacity()) *
           sizeof(ProductT<Word>);
  for (const auto &row : a.values)
    bytes += row.capacity() * sizeof(Word);
  for (const auto &ts : a.terms) {
    bytes += ts.capacity() * sizeof(TermT<Word>);
    for (const auto &t : ts)
      bytes += t.product.rows.capacity() * sizeof(Word);
  }
  for (const auto &ps : {&a.delta_products, &a.inverse_products})
    for (const auto &x : *ps)
      bytes += x.rows.capacity() * sizeof(Word);
  bytes += p.taylor_terms.capacity() * sizeof(p.taylor_terms[0]);
  for (const auto &ts : p.taylor_terms) {
    bytes += ts.capacity() * sizeof(TermT<Word>);
    for (const auto &t : ts)
      bytes += t.product.rows.capacity() * sizeof(Word);
  }
  p.owned_bytes = bytes;
}
template <class Plan>
std::unique_ptr<Plan> build_transform(unsigned s, uint32_t polynomial, uint32_t delta,
                                      const typename Plan::word_type *basis,
                                      size_t levels, bool base = false) {
  using Word = typename Plan::word_type;
  auto result = std::make_unique<Plan>();
  auto &a = result->transform;
  set_field(a.f, s, polynomial, delta, base);
  if (!basis || !levels || levels > (base ? s : 2 * s))
    throw std::runtime_error("invalid wide basis length");
  a.basis.assign(basis, basis + levels);
  a.terms.resize(levels);
  a.values.resize(levels);
  a.delta.resize(levels);
  a.inverse.resize(levels);
  a.delta_products.resize(levels);
  a.inverse_products.resize(levels);
  std::vector<Word> p{1};
  for (size_t k = 0; k < levels; k++) {
    if (!basis[k] || basis[k] >= a.f.q)
      throw std::runtime_error("wide basis range");
    for (size_t j = 0; j < p.size(); j++)
      if (p[j])
        a.terms[k].push_back({Word(1) << j, liftrs::constant_product(a.f, p[j])});
    a.values[k].resize(base ? s : 2 * s);
    for (unsigned j = 0; j < (base ? s : 2 * s); j++)
      a.values[k][j] = evaluate_linear(a.f, p, Word(1) << j);
    const auto d = evaluate_linear(a.f, p, basis[k]);
    if (!d)
      throw std::runtime_error("dependent wide basis");
    a.delta[k] = d;
    a.inverse[k] = liftrs::field_power(a.f, d, a.f.q - 2);
    if (a.f.mul(d, a.inverse[k]) != 1)
      throw std::runtime_error("invalid wide field inverse");
    a.delta_products[k] = liftrs::constant_product(a.f, d);
    a.inverse_products[k] = liftrs::constant_product(a.f, a.inverse[k]);
    std::vector<Word> next(p.size() + 1);
    for (size_t j = 0; j < p.size(); j++) {
      next[j + 1] ^= a.f.mul(p[j], p[j]);
      next[j] ^= a.f.mul(d, p[j]);
    }
    p = std::move(next);
  }
  account_plan(*result);
  return result;
}

template <class Matrix>
std::unique_ptr<Matrix> build_matrix(unsigned s, uint32_t polynomial, uint32_t delta,
                                     unsigned dim, unsigned split) {
  if (!split || split >= dim || dim > 2 * s)
    throw std::runtime_error("invalid matrix dimensions");
  using Plan = typename Matrix::plan_type;
  using Word = typename Matrix::word_type;
  auto m = std::make_unique<Matrix>();
  m->dim = dim;
  m->split = split;
  std::vector<Word> low(split + 1);
  for (unsigned j = 0; j <= split; j++)
    low[j] = Word(1) << j;
  m->rows = build_transform<Plan>(s, polynomial, delta, low.data(), low.size());
  auto &r = m->rows->transform;
  std::vector<Word> gamma(dim - split);
  for (unsigned j = 0; j < dim - split; j++)
    gamma[j] = r.offset(split, Word(1) << (split + j));
  m->columns = build_transform<Plan>(s, polynomial, delta, gamma.data(), gamma.size());
  m->rows->taylor_terms.resize(dim - split);
  for (unsigned k = 0; k < dim - split; k++)
    for (size_t j = 0; j + 1 < r.terms[split].size(); j++) {
      const auto &t = r.terms[split][j];
      auto c = t.product.coefficient;
      for (unsigned b = 0; b < k; b++)
        c = r.f.mul(c, c);
      m->rows->taylor_terms[k].push_back(
          {t.degree << k, liftrs::constant_product(r.f, c)});
    }
  account_plan(*m->rows);
  return m;
}
} // namespace liftrs::preparation
