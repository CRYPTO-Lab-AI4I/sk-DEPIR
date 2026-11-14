#pragma once
// Device execution only. Metadata and mathematical recurrences live in their
// shared headers; each launch assigns disjoint vectors or destinations.
#include "../liftrs-cpp/monomial.hpp"
#include "preparation_plan.hpp"
#include "transform.cuh"
namespace liftrs_cuda::preparation {
constexpr unsigned max_devices = 8;
// Uniform power-of-two slabs keep integer indexing exact at >2^32. Each GPU
// owns one slab; peer access is mandatory across every participating pair.
template <class Word> struct Shards {
  Word *data[max_devices]{};
  size_t chunk = 0;
  unsigned shift = 0;
  __device__ Word &operator[](size_t i) const {
    // Shapes and card counts are powers of two: avoid a64-bit division and
    // remainder on every coefficient access in the hot transform loops.
    return data[i >> shift][i & (chunk - 1)];
  }
};
template <class Word>
__global__ void transpose_scalar(Shards<Word> src, Word *dst, size_t first,
                                 size_t count, size_t rows, size_t columns) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j >= count)
    return;
  const size_t i = first + j;
  dst[j] = src[(i % rows) * columns + i / rows];
}
// Coalesced peer reads and local writes. Slab boundaries align with destination
// rows for ordinary sizes; tiny split-row slabs use the scalar fallback above.
template <class Word>
__global__ void transpose_tiles(Shards<Word> src, Word *dst, size_t first,
                                size_t count, size_t rows, size_t columns) {
  __shared__ Word data[32][33];
  const size_t row_tiles = (rows + 31) / 32, local_rows = count / rows;
  const size_t source_y = (size_t(blockIdx.x) % row_tiles) * 32,
               source_x = first / rows + (size_t(blockIdx.x) / row_tiles) * 32;
  const unsigned x = threadIdx.x, y = threadIdx.y;
  for (unsigned j = 0; j < 32; j += 8) {
    const size_t r = source_y + y + j, c = source_x + x;
    data[y + j][x] = (r < rows && c < columns && c < first / rows + local_rows)
                         ? src[r * columns + c]
                         : Word(0);
  }
  __syncthreads();
  for (unsigned j = 0; j < 32; j += 8) {
    const size_t r = source_x + y + j, c = source_y + x;
    if (c < rows && r < columns && r < first / rows + local_rows)
      dst[(r - first / rows) * rows + c] = data[x][y + j];
  }
}
template <class Word>
__global__ void vectors(Shards<Word> data, const Matrix<Word> *plan,
                        bool columns, bool inverse, size_t length, size_t jobs,
                        size_t first, size_t count, const uint32_t *logs,
                        const uint32_t *powers) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j >= count)
    return;
  const size_t vector = first + j;
  liftrs_cuda::vector_transform<Word>(data, *plan, columns, inverse, length,
                                      vector, jobs, vector, logs, powers, 0);
}
template <class Word>
__global__ void taylor(Shards<Word> data, const Matrix<Word> *plan, unsigned k,
                       bool quotient, size_t first, size_t count,
                       const uint32_t *logs, const uint32_t *powers) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j < count)
    liftrs_cuda::taylor_update<Word>(data, *plan, k, first + j, quotient, logs,
                                     powers);
}
template <class Word>
__global__ void pointwise(Word *a, const Word *b, size_t count, unsigned s,
                          uint32_t delta, const uint32_t *logs,
                          const uint32_t *powers) {
  const size_t i = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (i < count)
    a[i] = liftrs_cuda::multiply(a[i], b[i], s, delta, logs, powers);
}
template <class Word>
__global__ void widen(const uint32_t *input, Word *output, size_t count,
                      size_t available) {
  const size_t i = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (i < count)
    output[i] = i < available ? Word(input[i]) : Word(0);
}
template <class Word>
__global__ void narrow(const Word *input, uint32_t *output, size_t count,
                       unsigned s, uint32_t *bad) {
  const size_t i = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (i < count) {
    if (uint64_t(input[i]) >= (uint64_t(1) << s))
      atomicOr(bad, 1u);
    output[i] = uint32_t(input[i]);
  }
}
__global__ void weight_classes(unsigned char *out, size_t count, unsigned s,
                               unsigned t) {
  size_t i = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (i >= count)
    return;
  size_t v = i;
  unsigned weights[20];
  for (unsigned k = 0; k < s; k++) {
    weights[k] = unsigned(v % 3);
    v /= 3;
  }
  out[i] = liftrs::preparation::good_weights(weights, s, t);
}
__global__ void ternary_indices(uint32_t *out, size_t count, unsigned s) {
  size_t i = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (i >= count)
    return;
  uint32_t value = 0, power = 1;
  for (unsigned k = 0; k < s; k++, power *= 3)
    if ((i >> k) & 1)
      value += power;
  out[i] = value;
}
__global__ void expand_good(const unsigned char *classes,
                            const uint32_t *ternary, uint64_t *bits,
                            unsigned long long *total, size_t first,
                            size_t count, uint64_t n, unsigned s) {
  const size_t local = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  unsigned long long found = 0;
  if (local < count) {
    const size_t word = first + local;
    uint64_t value = 0;
    for (unsigned k = 0; k < 64 && 64 * word + k < n; k++) {
      const uint64_t index = 64 * word + k;
      const size_t x = index & ((uint64_t(1) << s) - 1), y = index >> s;
      if (classes[size_t(ternary[x]) + ternary[y]])
        value |= uint64_t(1) << k;
    }
    bits[local] = value;
    found = __popcll(value);
  }
  __shared__ unsigned long long sums[256];
  sums[threadIdx.x] = found;
  __syncthreads();
  for (unsigned stride = 128; stride; stride >>= 1) {
    if (threadIdx.x < stride)
      sums[threadIdx.x] += sums[threadIdx.x + stride];
    __syncthreads();
  }
  if (!threadIdx.x)
    atomicAdd(total, sums[0]);
}
template <class Word>
__global__ void root_factors(const int64_t *indices, size_t count,
                             uint32_t *output, uint64_t order, unsigned s,
                             uint32_t delta, Word omega, const Word *jumps,
                             const uint32_t *logs, const uint32_t *powers,
                             uint32_t *bad) {
  const size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  const unsigned lane = threadIdx.x & 31;
  Word factor = 1;
  if (j < count) {
    const uint64_t exponent = uint64_t(indices[j]);
    if (lane == 0)
      factor =
          liftrs_cuda::power(omega, order - exponent, s, delta, logs, powers);
    else {
      const uint64_t gap = exponent - uint64_t(indices[j - 1]);
      factor = gap < 64 ? jumps[gap]
                        : liftrs_cuda::power(omega, order - gap, s, delta, logs,
                                             powers);
    }
  }
  // Associative multiplicative scan produces omega^(-index) with only one
  // exponentiation per warp plus short-gap table lookups, including skipped
  // pairs.
  for (unsigned offset = 1; offset < 32; offset *= 2) {
    Word previous;
    if constexpr (sizeof(Word) == 8)
      previous = Word(__shfl_up_sync(
          0xffffffff, static_cast<unsigned long long>(factor), offset));
    else
      previous = __shfl_up_sync(0xffffffff, factor, offset);
    if (lane >= offset)
      factor = liftrs_cuda::multiply(factor, previous, s, delta, logs, powers);
  }
  if (j >= count)
    return;
  uint32_t *out = output + 4 * j;
  out[0] = out[1] = out[2] = out[3] = 0;
  const uint64_t exponent = uint64_t(indices[j]), q = uint64_t(1) << s,
                 partner = exponent * q % order;
  if (partner < exponent)
    return;
  const uint32_t a = uint32_t(factor) & uint32_t(q - 1),
                 b = uint32_t(factor >> s);
  if (partner == exponent) {
    if (b)
      atomicOr(bad, 1u);
    out[0] = a;
    out[1] = 1;
    out[3] = 1;
  } else {
    out[0] =
        liftrs_cuda::base_product(a, a, logs, powers) ^
        liftrs_cuda::base_product(a, b, logs, powers) ^
        liftrs_cuda::base_product(
            delta, liftrs_cuda::base_product(b, b, logs, powers), logs, powers);
    out[1] = b;
    out[2] = 1;
    out[3] = 2;
  }
}

} // namespace liftrs_cuda::preparation
