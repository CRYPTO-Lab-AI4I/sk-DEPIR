#pragma once
// Portable polynomial-basis kernels matching liftrs-go. Tables and transform
// constants are trusted experiment fixtures, not a public cache format. There is
// no cryptographic stage or application encoder here; arithmetic is not constant-time.
#include <algorithm>
#include <cstdint>
#include <stdexcept>
#include <type_traits>
#include <vector>
namespace liftrs {
template <class Word> struct FieldT {
  static_assert(std::is_same_v<Word, uint32_t> || std::is_same_v<Word, uint64_t>);
  uint32_t s = 13, polynomial = 0x201b, delta = 1;
  uint64_t q =
      uint64_t(1)
      << 26; // Cardinality needs64 bits even for the32-bit value specialization.
  uint32_t base_mul(uint32_t a, uint32_t b) const {
    uint32_t r = 0;
    while (b) {
      if (b & 1)
        r ^= a;
      b >>= 1;
      a <<= 1;
      if (a & (1u << s))
        a ^= polynomial;
    }
    return r;
  }
  Word mul(Word a, Word b) const {
    if (q == (1u << s))
      return base_mul(a, b);
    const uint32_t mask = (1u << s) - 1, aa = a & mask, ab = a >> s, ba = b & mask,
                   bb = b >> s;
    const uint32_t ac = base_mul(aa, ba), bd = base_mul(ab, bb);
    return Word(ac ^ base_mul(delta, bd)) |
           (Word(base_mul(aa ^ ab, ba ^ bb) ^ ac) << s);
  }
};
template <class Word> struct ProductT {
  Word coefficient = 0;
  std::vector<Word> rows;
  Word mul(const FieldT<Word> &f, Word x) const {
    if (coefficient == 0)
      return 0;
    if (coefficient == 1)
      return x;
    if (rows.empty())
      return f.mul(coefficient, x);
    Word y = 0;
    for (size_t j = 0; j < rows.size(); j += 16) {
      y ^= rows[j + (x & 15)];
      x >>= 4;
    }
    return y;
  }
};
template <class Word> struct TermT {
  Word degree;
  ProductT<Word> product;
};
template <class Word> struct TransformT {
  FieldT<Word> f;
  std::vector<std::vector<TermT<Word>>> terms;
  std::vector<std::vector<Word>> values;
  std::vector<Word>
      basis; // Ordered field coordinates; empty means canonical powers of two.
  std::vector<Word> delta, inverse;
  std::vector<ProductT<Word>> delta_products, inverse_products;
  Word offset(unsigned k, Word x) const {
    Word out = 0;
    while (x) {
      unsigned j;
      if constexpr (sizeof(Word) == 8)
        j = __builtin_ctzll(x);
      else
        j = __builtin_ctz(x);
      out ^= values[k][j];
      x &= x - 1;
    }
    return out;
  }
  void recurse(Word *data, size_t length, unsigned level, Word start,
               bool invert) const {
    if (!level)
      return;
    const unsigned k = level - 1;
    const size_t half = length / 2;
    const Word c = offset(k, start);
    const auto &ts = terms[k];
    if (!invert) {
      for (size_t i = length; i-- > half;) {
        Word v = data[i];
        for (size_t j = 0; j + 1 < ts.size(); j++)
          data[i - half + ts[j].degree] ^= ts[j].product.mul(f, v);
      }
      for (size_t j = 0; j < half; j++) {
        Word v = data[half + j], left = data[j] ^ f.mul(c, v);
        data[j] = left;
        data[half + j] = left ^ delta_products[k].mul(f, v);
      }
    }
    recurse(data, half, k, start, invert);
    recurse(data + half, half, k, start ^ (basis.empty() ? Word(half) : basis[k]),
            invert);
    if (invert) {
      for (size_t j = 0; j < half; j++) {
        Word v = inverse_products[k].mul(f, data[j] ^ data[half + j]);
        data[j] ^= f.mul(c, v);
        data[half + j] = v;
      }
      for (size_t j = 0; j < half; j++) {
        Word v = data[half + j];
        for (size_t n = 0; n + 1 < ts.size(); n++)
          data[j + ts[n].degree] ^= ts[n].product.mul(f, v);
      }
    }
  }
  // Borrow a pointer-free caller buffer for this call; no allocation or retained
  // message pointer. The immutable plan is shared across independent encodings.
  void apply(Word *data, size_t length, bool invert = false, Word start = 0) const {
    if (!data || !length || (length & (length - 1)) || length > f.q || start >= f.q)
      throw std::runtime_error("invalid transform length");
    unsigned level = 0;
    for (size_t n = length; n > 1; n >>= 1)
      level++;
    if (level > terms.size())
      throw std::runtime_error("transform exceeds basis");
    recurse(data, length, level, start, invert);
  }
  void apply(std::vector<Word> &data, bool invert = false) const {
    apply(data.data(), data.size(), invert);
  }
};
// Word-width aliases preserve the existing32-bit ABI. Both specializations use
// the same arithmetic/transform implementation and shared online schedule.
using Field = FieldT<uint32_t>;
using Product = ProductT<uint32_t>;
using Term = TermT<uint32_t>;
using Transform = TransformT<uint32_t>;
using Field64 = FieldT<uint64_t>;
using Product64 = ProductT<uint64_t>;
using Term64 = TermT<uint64_t>;
using Transform64 = TransformT<uint64_t>;
#if defined(__GNUC__) || defined(__clang__)
__attribute__((noinline))
#endif
inline void
batch(const std::vector<Product> &products, const Field &f,
      const std::vector<uint32_t> &in, std::vector<uint32_t> &out, bool rotating) {
  for (size_t start = 0; start < in.size(); start += 4096) {
    size_t end = std::min(start + 4096, in.size());
    size_t index = rotating ? (start / 4096) % products.size() : products.size() / 2;
    const auto &p = products[index];
    for (size_t i = start; i < end; i++)
      out[i] = p.mul(f, in[i]);
  }
}
} // namespace liftrs
