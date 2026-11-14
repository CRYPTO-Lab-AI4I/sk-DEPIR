#pragma once
// Arithmetic and per-vector recurrences shared by legacy client kernels and
// width-selected server preparation. No backend chooses another memory layout.
#include <cstddef>
#include <cstdint>
#if defined(__CUDACC__)
#include <cuda_runtime.h>
#endif

namespace liftrs_cuda {
__host__ __device__ inline uint32_t base_product(uint32_t a, uint32_t b,
                                                 const uint32_t *log,
                                                 const uint32_t *powers) {
  return a && b ? powers[log[a] + log[b]] : 0;
}
template <class Word>
__host__ __device__ Word multiply(Word a, Word b, unsigned s, uint32_t delta,
                                  const uint32_t *log, const uint32_t *powers) {
  if (!a || !b)
    return 0;
  if (a == 1)
    return b;
  if (b == 1)
    return a;
  const uint32_t mask = (1u << s) - 1, aa = uint32_t(a) & mask,
                 ab = uint32_t(a >> s), ba = uint32_t(b) & mask,
                 bb = uint32_t(b >> s);
  const uint32_t ac = base_product(aa, ba, log, powers),
                 bd = base_product(ab, bb, log, powers);
  return Word(ac ^ base_product(delta, bd, log, powers)) |
         (Word(base_product(aa ^ ab, ba ^ bb, log, powers) ^ ac) << s);
}
template <class Word>
__host__ __device__ Word power(Word x, uint64_t exponent, unsigned s,
                               uint32_t delta, const uint32_t *log,
                               const uint32_t *powers) {
  Word out = 1;
  while (exponent) {
    if (exponent & 1)
      out = multiply(out, x, s, delta, log, powers);
    exponent >>= 1;
    if (exponent)
      x = multiply(x, x, s, delta, log, powers);
  }
  return out;
}

// Accessor returns a mutable coefficient reference. It can be a local pointer
// or a peer-enabled sharded array; each vector has a single owning CUDA thread.
template <class Word, class Accessor, class Metadata>
__device__ __forceinline__ void
vector_transform(Accessor data, const Metadata &m, bool columns, bool inverse,
                 size_t length, size_t tile, size_t stride, size_t offset,
                 const uint32_t *log, const uint32_t *powers,
                 size_t first_vector) {
  const auto &p = columns ? m.columns : m.rows;
  const unsigned level = columns ? m.dimension - m.split : m.split;
  for (unsigned stage = 0; stage < level; stage++) {
    const unsigned k = inverse ? stage : level - 1 - stage;
    const size_t size = size_t(1) << (k + 1), half = size / 2;
    for (size_t group = 0; group < length; group += size) {
      Word point = columns ? Word(0) : Word((tile + first_vector) * length);
      for (unsigned bit = 0; bit < level; bit++)
        if ((group >> bit) & 1)
          point ^= p.basis[bit];
      Word c = 0;
      for (unsigned bit = 0; bit < 2 * m.s; bit++)
        if ((point >> bit) & 1)
          c ^= p.values[k][bit];
      if (!inverse) {
        for (size_t i = size; i-- > half;) {
          const Word z = data[(group + i) * stride + offset];
          for (unsigned t = 0; t + 1 < p.count[k]; t++)
            data[(group + i - half + p.degree[k][t]) * stride + offset] ^=
                multiply(p.coef[k][t], z, m.s, m.delta, log, powers);
        }
        for (size_t j = 0; j < half; j++) {
          const size_t a = (group + j) * stride + offset,
                       b = (group + half + j) * stride + offset;
          const Word v = data[b],
                     left = data[a] ^ multiply(c, v, m.s, m.delta, log, powers);
          data[a] = left;
          data[b] = left ^ multiply(p.delta[k], v, m.s, m.delta, log, powers);
        }
      } else {
        for (size_t j = 0; j < half; j++) {
          const size_t a = (group + j) * stride + offset,
                       b = (group + half + j) * stride + offset;
          const Word v = multiply(p.inverse[k], Word(data[a] ^ data[b]), m.s,
                                  m.delta, log, powers);
          data[a] ^= multiply(c, v, m.s, m.delta, log, powers);
          data[b] = v;
        }
        for (size_t j = 0; j < half; j++) {
          const Word v = data[(group + half + j) * stride + offset];
          for (unsigned t = 0; t + 1 < p.count[k]; t++)
            data[(group + j + p.degree[k][t]) * stride + offset] ^=
                multiply(p.coef[k][t], v, m.s, m.delta, log, powers);
        }
      }
    }
  }
}

// The quotient reads only the untouched upper quarter; the remainder reads the
// completed quotient. Synchronize ALL devices between these two passes.
template <class Word, class Accessor, class Metadata>
__device__ __forceinline__ void
taylor_update(Accessor x, const Metadata &m, unsigned k, size_t index,
              bool quotient, const uint32_t *log, const uint32_t *powers) {
  const auto &p = m.taylor[k];
  const size_t half = size_t(1) << (m.split + k), j = index % half,
               offset = (index / half) * 2 * half;
  if (quotient) {
    if (j >= half / 2)
      return;
    Word value = x[offset + half + j];
    for (unsigned t = 0; t < p.count; t++) {
      const auto d = p.degree[t];
      if (d > j)
        value ^= multiply(p.coef[t], Word(x[offset + 2 * half + j - d]), m.s,
                          m.delta, log, powers);
    }
    x[offset + half + j] = value;
  } else {
    Word value = x[offset + j];
    for (unsigned t = 0; t < p.count; t++) {
      const auto d = p.degree[t];
      if (d <= j)
        value ^= multiply(p.coef[t], Word(x[offset + half + j - d]), m.s,
                          m.delta, log, powers);
    }
    x[offset + j] = value;
  }
}
} // namespace liftrs_cuda
