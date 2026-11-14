#pragma once
// Host metadata only, constructed once with the common native sparse builder.
#include "../liftrs-cpp/prepare_plan.hpp"
namespace liftrs_cuda::preparation {
constexpr unsigned max_levels = 21, max_bits = 40;
template <class Word> struct Tile {
  unsigned count[max_levels]{};
  Word degree[max_levels][max_levels]{}, coef[max_levels][max_levels]{},
      values[max_levels][max_bits]{}, basis[max_levels]{}, delta[max_levels]{},
      inverse[max_levels]{};
};
template <class Word> struct Matrix {
  Tile<Word> rows, columns;
  struct Taylor {
    unsigned count = 0;
    Word degree[max_levels]{}, coef[max_levels]{};
  } taylor[max_levels];
  unsigned split = 0, dimension = 0, s = 0;
  uint32_t delta = 0;
};
template <class Word> Tile<Word> tile(const liftrs::TransformT<Word> &plan) {
  Tile<Word> out{};
  if (plan.basis.size() > max_levels)
    throw std::runtime_error("GPU plan level bound");
  for (size_t k = 0; k < plan.basis.size(); k++) {
    out.basis[k] = plan.basis[k];
    out.count[k] = unsigned(plan.terms[k].size());
    if (out.count[k] > max_levels)
      throw std::runtime_error("GPU term bound");
    for (size_t j = 0; j < plan.terms[k].size(); j++) {
      out.degree[k][j] = plan.terms[k][j].degree;
      out.coef[k][j] = plan.terms[k][j].product.coefficient;
    }
    std::copy(plan.values[k].begin(), plan.values[k].end(), out.values[k]);
    out.delta[k] = plan.delta[k];
    out.inverse[k] = plan.inverse[k];
  }
  return out;
}
template <class Word>
Matrix<Word> matrix(unsigned s, uint32_t polynomial, uint32_t delta, size_t n,
                    bool base = false) {
  Matrix<Word> out{};
  out.s = s;
  out.delta = delta;
  if (base) {
    std::vector<Word> basis(s);
    for (unsigned i = 0; i < s; i++)
      basis[i] = Word(1) << i;
    auto p = liftrs::preparation::build_transform<liftrs::TransformPlan<Word>>(
        s, polynomial, delta, basis.data(), s, true);
    out.split = s;
    out.dimension = 2 * s;
    out.rows = out.columns = tile(p->transform);
    return out;
  }
  unsigned dim = 0;
  for (auto v = n; v > 1; v >>= 1)
    dim++;
  auto p = liftrs::preparation::build_matrix<liftrs::MatrixPlan<Word>>(
      s, polynomial, delta, dim, dim / 2);
  out.split = p->split;
  out.dimension = dim;
  out.rows = tile(p->rows->transform);
  out.columns = tile(p->columns->transform);
  for (size_t k = 0; k < p->rows->taylor_terms.size(); k++) {
    const auto &terms = p->rows->taylor_terms[k];
    out.taylor[k].count = unsigned(terms.size());
    if (terms.size() > max_levels)
      throw std::runtime_error("GPU Taylor term bound");
    for (size_t j = 0; j < terms.size(); j++) {
      out.taylor[k].degree[j] = terms[j].degree;
      out.taylor[k].coef[j] = terms[j].product.coefficient;
    }
  }
  return out;
}
} // namespace liftrs_cuda::preparation
