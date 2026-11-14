#pragma once
// Internal immutable transform metadata. Native import and sparse construction
// share this representation; neither owns or retains caller data buffers.
#include "transform.hpp"
#include <memory>

namespace liftrs {
template <class Word> struct TransformPlan {
  using word_type = Word;
  TransformT<Word> transform;
  std::vector<std::vector<TermT<Word>>> taylor_terms;
  size_t owned_bytes = 0;
};
template <class Word, class Plan = TransformPlan<Word>> struct MatrixPlan {
  using plan_type = Plan;
  using word_type = Word;
  std::unique_ptr<Plan> rows, columns;
  unsigned dim = 0, split = 0;
};
} // namespace liftrs

// Opaque legacy C ABI handle retains its original layout and word width.
struct liftrs_plan : liftrs::TransformPlan<uint32_t> {};
