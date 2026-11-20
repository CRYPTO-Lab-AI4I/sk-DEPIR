#pragma once
// Included inside the preparation executor namespace; shared bounded source view.
// Each source window is a snapshot: quotient sources lie in the untouched
// upper quarter, and remainder sources in the completed upper half. Distinct
// target tiles never feed each other within a pass. Overlapping source windows
// therefore contain identical values; the first matching window is sufficient.
template <class Word> struct TaylorWindow {
  Word *data;
  const Matrix<Word> *matrix;
  size_t block, half, first, count;
  unsigned level;
  bool quotient;
  __device__ Word &operator[](size_t index) const {
    const size_t target = block + (quotient ? half : 0) + first;
    if (index >= target && index - target < count)
      return data[index - target];
    const auto &terms = matrix->taylor[level];
    for (unsigned t = 0; t < terms.count; t++) {
      const size_t source = block + (quotient ? 2 * half : half) + first - size_t(terms.degree[t]);
      if (index >= source && index - source < count)
        return data[(t + 1) * count + index - source];
    }
    // A missing source would silently corrupt arithmetic; fail the kernel instead.
    asm("trap;");
    return data[0];
  }
};
template <class Word>
__global__ void window_taylor(TaylorWindow<Word> view, size_t first_index,
                               const uint32_t *logs, const uint32_t *powers) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j < view.count)
    liftrs_cuda::taylor_update<Word>(view, *view.matrix, view.level,
                                    first_index + j, view.quotient, logs, powers);
}
