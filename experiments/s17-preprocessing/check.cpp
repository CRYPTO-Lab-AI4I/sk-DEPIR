// go:build ignore

// Exercise BOTH storage specializations on the same complete small domain.
// Independent schoolbook/Horner checks prevent a paired forward/inverse bug
// from appearing correct; the Go suite independently covers s17 arithmetic.
#include "preparation.hpp"
#include <iostream>
using namespace liftrs;
void require(bool condition, const char *message) {
  if (!condition)
    throw std::runtime_error(message);
}
uint32_t multiply(uint32_t a, uint32_t b) {
  uint32_t result = 0;
  while (b) {
    if (b & 1)
      result ^= a;
    b >>= 1;
    a <<= 1;
    if (a & 4)
      a ^= 7;
  }
  return result;
}
uint64_t extension(uint64_t a, uint64_t b) {
  auto aa = a & 3, ab = a >> 2, ba = b & 3, bb = b >> 2;
  auto ac = multiply(aa, ba), bd = multiply(ab, bb);
  return (ac ^ multiply(2, bd)) |
         (uint64_t(multiply(aa ^ ab, ba ^ bb) ^ ac) << 2);
}
template <class Word> void check(unsigned workers) {
  preparation::Executor<Word> executor(2, 7, 2, 4, workers, 16, 8 << 20);
  std::vector<Word> input(16);
  for (size_t i = 0; i < 16; i++)
    input[i] = (i * 7 + 1) & 15;
  auto value = input;
  executor.apply(value.data(), 16, false);
  for (size_t x = 0; x < 16; x++) {
    uint64_t expected = 0;
    for (size_t j = 16; j-- > 0;)
      expected = extension(expected, x) ^ input[j];
    require(value[x] == expected, "Horner");
  }
  executor.apply(value.data(), 16, true);
  require(value == input, "inverse");
  std::vector<uint32_t> a{1, 2, 3, 2, 1}, b{3, 2, 1, 3, 2}, out(9), expected(9);
  executor.product(a.data(), a.size(), b.data(), b.size(), out.data(), 9, 16);
  for (size_t i = 0; i < a.size(); i++)
    for (size_t j = 0; j < b.size(); j++)
      expected[i + j] ^= multiply(a[i], b[j]);
  require(out == expected, "product");
  // m2 tensor evaluation is an independent pair of nested Horner evaluations.
  std::vector<Word> grid(16);
  for (size_t i = 0; i < 16; i++)
    grid[i] = (3 * i + 1) & 3;
  auto coefficients = grid;
  executor.base_grid(grid.data(), 16);
  for (size_t y = 0; y < 4; y++)
    for (size_t x = 0; x < 4; x++) {
      uint32_t sum = 0;
      for (size_t row = 4; row-- > 0;) {
        uint32_t v = 0;
        for (size_t col = 4; col-- > 0;)
          v = multiply(v, x) ^ coefficients[row * 4 + col];
        sum = multiply(sum, y) ^ v;
      }
      require(grid[y * 4 + x] == sum, "tensor Horner");
    }
  executor.apply(grid.data(), 16, true);
  const auto count = executor.count_good(1);
  std::vector<int64_t> good(count);
  executor.fill_good(good.data(), count);
  // Enumerate every permitted Algorithm1 sum, including the zero-only case.
  for (unsigned index = 0; index < 16; index++) {
    bool possible[7] = {true};
    for (unsigned bit = 0; bit < 2; bit++) {
      bool next[7] = {false};
      auto weight = ((index >> bit) & 1) + ((index >> (2 + bit)) & 1);
      for (unsigned v = 0; v < 7; v++)
        if (possible[v])
          for (unsigned k = 0; k <= weight; k++)
            if (v + (k << bit) < 7)
              next[v + (k << bit)] = true;
      std::copy(next, next + 7, possible);
    }
    bool expectedGood = !possible[3] && !possible[6];
    require(std::binary_search(good.begin(), good.end(), index) == expectedGood,
            "exact good set");
  }
  auto before = executor.bytes();
  executor.release_products();
  require(executor.bytes() < before, "reuse release");
  std::cout << "{\"word_bytes\":" << sizeof(Word) << ",\"workers\":" << workers
            << ",\"complete_domain\":16,\"passed\":true}\n";
}
int main() {
  try {
    for (unsigned workers : {1, 4}) {
      check<uint32_t>(workers);
      check<uint64_t>(workers);
    }
    return 0;
  } catch (const std::exception &e) {
    std::cerr << e.what() << '\n';
    return 1;
  }
}
