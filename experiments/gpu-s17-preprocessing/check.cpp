// go:build ignore

// Host execution of the SAME CUDA recurrences, across simulated peer slabs.
// Native bit-polynomial multiplication supplies an independent Horner oracle.
#define __host__
#define __device__
#define __forceinline__ inline
#include "../liftrs-cpp/progress.hpp"
#include "preparation_plan.hpp"
#include "transform.cuh"
#include <iostream>
using namespace liftrs;
using namespace liftrs_cuda::preparation;
void require(bool value, const char *message) {
  if (!value)
    throw std::runtime_error(message);
}
template <class Word> struct View {
  std::vector<std::vector<Word>> &slabs;
  size_t chunk;
  Word &operator[](size_t index) const {
    return slabs[index / chunk][index % chunk];
  }
};
template <class Word>
void check(unsigned s, uint32_t polynomial, unsigned count) {
  const size_t n = 128;
  FieldT<Word> field;
  field.s = s;
  field.polynomial = polynomial;
  field.q = uint64_t(1) << (2 * s);
  // Choose delta with trace1, as the common scalar tower does.
  FieldT<Word> base = field;
  base.q = uint64_t(1) << s;
  for (uint32_t d = 1;; d++) {
    Word trace = 0, x = d;
    for (unsigned j = 0; j < s; j++) {
      trace ^= x;
      x = base.mul(x, x);
    }
    if (trace == 1) {
      field.delta = d;
      break;
    }
  }
  auto metadata = matrix<Word>(s, polynomial, field.delta, n);
  if (s == 17) {
    auto full = matrix<Word>(s, polynomial, field.delta, size_t(1) << 34);
    require(full.dimension == 34 && full.split == 17,
            "full-domain dimensions narrowed");
    require(full.rows.values[0][33] == (uint64_t(1) << 33),
            "full-domain offset constant narrowed");
  }
  std::vector<uint32_t> logs(size_t(1) << s),
      powers(2 * ((size_t(1) << s) - 1));
  Word generator = 2;
  while (!is_primitive(base, generator))
    generator++;
  Word current = 1;
  const size_t order = (size_t(1) << s) - 1;
  for (size_t j = 0; j < order; j++) {
    logs[size_t(current)] = uint32_t(j);
    powers[j] = uint32_t(current);
    current = base.mul(current, generator);
  }
  std::copy(powers.begin(), powers.begin() + order, powers.begin() + order);
  std::vector<std::vector<Word>> storage(count, std::vector<Word>(n / count)),
      scratch(count, std::vector<Word>(n / count));
  View<Word> x{storage, n / count}, tmp{scratch, n / count};
  std::vector<Word> original(n);
  for (size_t j = 0; j < n; j++)
    original[j] = x[j] = Word((j * 19097 + 1) ^ (uint64_t(1) << (2 * s - 1)));
  auto transform = [&](bool inverse) {
    auto convert = [&] {
      for (unsigned stage = 0; stage < metadata.dimension - metadata.split;
           stage++) {
        unsigned k =
            inverse ? stage : metadata.dimension - metadata.split - 1 - stage;
        for (unsigned pass = 0; pass < 2; pass++)
          for (size_t j = 0; j < n / 2; j++)
            liftrs_cuda::taylor_update<Word>(x, metadata, k, j,
                                             inverse ? pass != 0 : pass == 0,
                                             logs.data(), powers.data());
      }
    };
    const size_t length = size_t(1) << metadata.split, jobs = n / length;
    auto columns = [&] {
      for (size_t j = 0; j < length; j++)
        liftrs_cuda::vector_transform<Word>(x, metadata, true, inverse, jobs, j,
                                            length, j, logs.data(),
                                            powers.data(), 0);
    };
    auto rows = [&] {
      for (size_t j = 0; j < n; j++)
        tmp[j] = x[(j % jobs) * length + j / jobs];
      for (size_t j = 0; j < jobs; j++)
        liftrs_cuda::vector_transform<Word>(tmp, metadata, false, inverse,
                                            length, j, jobs, j, logs.data(),
                                            powers.data(), 0);
      for (size_t j = 0; j < n; j++)
        x[j] = tmp[(j % length) * jobs + j / length];
    };
    if (!inverse) {
      convert();
      columns();
      rows();
    } else {
      rows();
      columns();
      convert();
    }
  };
  // Use only canonical values for smaller fields; retain >32-bit witnesses s17.
  for (size_t j = 0; j < n; j++)
    original[j] = x[j] = original[j] & Word(field.q - 1);
  transform(false);
  for (size_t j = 0; j < n; j++) {
    Word expected = 0;
    for (size_t k = n; k-- > 0;)
      expected = field.mul(expected, Word(j)) ^ original[k];
    require(x[j] == expected, "shared CUDA Horner recurrence");
  }
  transform(true);
  for (size_t j = 0; j < n; j++)
    require(x[j] == original[j], "shared CUDA inverse recurrence");
  std::cout << "{\"s\":" << s << ",\"word_bytes\":" << sizeof(Word)
            << ",\"simulated_devices\":" << count << ",\"match\":true}\n";
}
int main() {
  try {
    for (unsigned count : {1, 2, 4}) {
      check<uint32_t>(4, 0x13, count);
      check<uint64_t>(4, 0x13, count);
      check<uint64_t>(17, 0x20009, count);
    }
    liftrs::execution::progress_record("logging-self-check", "recurrence-cases",
                                       "completed", 9, 9, 0, "cases");
    liftrs::execution::progress_record(
        "legacy-log-self-check", "recurrence-cases", "completed", 9, 9, 0);
    return 0;
  } catch (const std::exception &e) {
    std::cerr << e.what() << '\n';
    return 1;
  }
}
