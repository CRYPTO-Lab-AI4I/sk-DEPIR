#pragma once
// Client-only movement/assembly kernels. Field multiplication, transforms and
// distributed array addressing are shared with the resident executor.
#include "preparation_kernels.cuh"
namespace liftrs_cuda::encoding {
using preparation::Shards;
constexpr size_t canonical_chunk = 4096;

// Prefix H is distributed in shorter slabs. Write padded full slabs into dead
// scratch, allowing the full transform to use the old prefix arena as scratch.
template <class Word>
__global__ void promote(Shards<Word> prefix, Word *out, size_t first, size_t count,
                        size_t k) {
  size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j < count) {
    size_t i = first + j;
    out[j] = i < k ? prefix[i] : 0;
  }
}

template <class Word>
__device__ Word first_position(Word omega, uint64_t index, unsigned s, uint32_t delta,
                               const uint32_t *logs, const uint32_t *powers) {
  return liftrs_cuda::power(omega, index + 1, s, delta, logs, powers);
}

template <class Word>
__global__ void canonical(const Word *coefficients, Shards<uint32_t> output,
                          size_t first, size_t count, size_t n, Word omega, unsigned s,
                          uint32_t delta, const uint32_t *logs, const uint32_t *powers,
                          uint32_t *parity, uint32_t *bad, size_t first_job,
                          size_t jobs) {
  constexpr size_t chunk = canonical_chunk;
  size_t local = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (local >= jobs)
    return;
  size_t job = first_job + local, begin = job * chunk;
  if (begin >= count)
    return;
  Word point = first_position(omega, uint64_t(first + begin), s, delta, logs, powers);
  uint32_t sum = 0;
  for (size_t j = begin;
       j < (count < begin + chunk ? count : begin + chunk) && first + j < n - 1; j++) {
    Word value = coefficients[j];
    if (value >= (Word(1) << s))
      atomicOr(bad, 1u);
    else {
      output[size_t(point)] = uint32_t(value);
      sum ^= uint32_t(value);
    }
    point = liftrs_cuda::multiply(point, omega, s, delta, logs, powers);
  }
  parity[job] = sum;
}

// Per-card XOR reduces small chunk results; the host combines one public parity
// symbol per card, then writes canonical coordinate zero after all cards join.
__global__ void reduce_parity(uint32_t *values, size_t jobs) {
  __shared__ uint32_t sums[256];
  uint32_t sum = 0;
  for (size_t i = threadIdx.x; i < jobs; i += blockDim.x)
    sum ^= values[i];
  sums[threadIdx.x] = sum;
  __syncthreads();
  for (unsigned offset = 128; offset; offset >>= 1) {
    if (threadIdx.x < offset)
      sums[threadIdx.x] ^= sums[threadIdx.x + offset];
    __syncthreads();
  }
  if (threadIdx.x == 0)
    values[jobs] = sums[0];
}
} // namespace liftrs_cuda::encoding
