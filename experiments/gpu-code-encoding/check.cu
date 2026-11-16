// Direct checker instantiates both workspace widths on genuine small codes.
// It includes the same executor and client implementation, never copied kernels.
#include "../../encoding/liftrs-cuda/preparation.cu"
#include <fstream>
#include <iostream>
#include <limits>

struct Reader {
  std::ifstream file;
  explicit Reader(const char *name) : file(name, std::ios::binary) {
    if (!file)
      throw std::runtime_error("fixture open");
  }
  template <class T> T scalar() {
    T value = 0;
    for (unsigned i = 0; i < sizeof(T); i++) {
      int b = file.get();
      if (b == EOF)
        throw std::runtime_error("fixture truncated");
      value |= T(static_cast<unsigned char>(b)) << (8 * i);
    }
    return value;
  }
  template <class T> std::vector<T> array(size_t n) {
    std::vector<T> out(n);
    for (auto &v : out)
      v = scalar<T>();
    return out;
  }
};
void require(bool condition, const char *message) {
  if (!condition)
    throw std::runtime_error(message);
}
struct Fixture {
  unsigned s;
  uint32_t polynomial, delta;
  uint64_t omega, k, n;
  std::vector<uint32_t> logs, powers, p, message, word;
  std::vector<uint64_t> spectrum;
  explicit Fixture(Reader &r) {
    s = r.scalar<uint32_t>();
    require(s >= 3 && s <= 8, "fixture degree bound");
    polynomial = r.scalar<uint32_t>();
    delta = r.scalar<uint32_t>();
    omega = r.scalar<uint64_t>();
    k = r.scalar<uint64_t>();
    n = r.scalar<uint64_t>();
    require(n == (uint64_t(1) << (2 * s)) && k > 0 && k <= n / 2, "fixture shape");
    logs = r.array<uint32_t>(size_t(1) << s);
    powers = r.array<uint32_t>(2 * ((size_t(1) << s) - 1));
    p = r.array<uint32_t>(k + 1);
    spectrum = r.array<uint64_t>(n);
    message = r.array<uint32_t>(k);
    word = r.array<uint32_t>(n);
  }
};
template <class Word> void verify(const Fixture &f, unsigned count, int layout) {
  const auto began = std::chrono::steady_clock::now();
  const std::string label = "GPU-code-s" + std::to_string(f.s) + "-k" +
                            std::to_string(f.k) + "-w" + std::to_string(sizeof(Word)) +
                            "-g" + std::to_string(count) + "-layout" +
                            std::to_string(layout);
  liftrs::execution::progress_record(label.c_str(), "cases", "started", 0, 1, 0);
  std::vector<int> devices(count);
  for (unsigned i = 0; i < count; i++)
    devices[i] = int(i);
  std::vector<Word> spectrum(f.spectrum.begin(), f.spectrum.end());
  Resident<Word> engine(f.s, f.polynomial, f.delta, f.omega, f.logs.data(),
                        f.logs.size(), f.powers.data(), f.powers.size(), devices.data(),
                        count, f.n, 1ull << 30);
  engine.set_code(f.p.data(), f.p.size(), spectrum.data(), f.n, layout);
  const auto owned = engine.owned();
  require(owned > 0 && owned <= (uint64_t(count) << 30), "owned bytes cap");
  std::vector<uint32_t> out(f.n + 2, 0xdeadbeef);
  double stats[10];
  auto encode = [&](const std::vector<uint32_t> &message) {
    engine.encode_code(message.data(), message.size(), out.data() + 1, f.n, stats);
    require(out.front() == 0xdeadbeef && out.back() == 0xdeadbeef,
            "host output canary");
  };
  encode(f.message);
  require(std::equal(f.word.begin(), f.word.end(), out.begin() + 1),
          "independent Go codeword");
  auto bad = f.message;
  bad[0] = uint32_t(1) << f.s;
  bool rejected = false;
  try {
    encode(bad);
  } catch (const std::exception &) {
    rejected = true;
  }
  require(rejected, "invalid message accepted");
  rejected = false;
  try {
    engine.encode_code(f.message.data(), f.k, const_cast<uint32_t *>(f.message.data()),
                       f.n, stats);
  } catch (const std::exception &) {
    rejected = true;
  }
  require(rejected, "overlap accepted");
  std::vector<uint32_t> zero(f.k);
  encode(zero);
  require(std::all_of(out.begin() + 1, out.end() - 1, [](auto x) { return x == 0; }),
          "zero message");
  encode(f.message);
  require(std::equal(f.word.begin(), f.word.end(), out.begin() + 1),
          "recovery/immutable constants");
  require(engine.owned() == owned, "repeat device allocations");
  rejected = false;
  try {
    engine.set_code(f.p.data(), f.p.size(), spectrum.data(), f.n, layout);
  } catch (const std::exception &) {
    rejected = true;
  }
  require(rejected, "second configure accepted");
  liftrs::execution::progress_record(
      label.c_str(), "cases", "completed", 1, 1,
      std::chrono::duration_cast<std::chrono::milliseconds>(
          std::chrono::steady_clock::now() - began)
          .count());
}

__global__ void positions(const uint64_t *indices, uint64_t *output, size_t count,
                          uint64_t omega, uint32_t delta, const uint32_t *logs,
                          const uint32_t *powers) {
  size_t j = size_t(blockIdx.x) * blockDim.x + threadIdx.x;
  if (j >= count)
    return;
  uint64_t point =
      liftrs_cuda::encoding::first_position(omega, indices[j], 17, delta, logs, powers);
  output[3 * j] = point;
  // The same slab division/mask as Shards::operator[], including shifts>=32.
  output[3 * j + 1] = point >> 32;
  output[3 * j + 2] = point & ((uint64_t(1) << 32) - 1);
}
void wide_positions(Reader &r) {
  auto poly = r.scalar<uint32_t>(), delta = r.scalar<uint32_t>();
  auto omega = r.scalar<uint64_t>();
  (void)poly;
  auto logs = r.array<uint32_t>(1 << 17),
       powers = r.array<uint32_t>(2 * ((1 << 17) - 1));
  auto count = r.scalar<uint32_t>();
  require(count > 0 && count <= 32, "wide coordinate count");
  std::vector<uint64_t> indices(count), expected(count), actual(count * 3);
  for (unsigned i = 0; i < count; i++) {
    indices[i] = r.scalar<uint64_t>();
    expected[i] = r.scalar<uint64_t>();
  }
  Buffer dl, dp, di, dout;
  dl.allocate(0, logs.size() * 4);
  dp.allocate(0, powers.size() * 4);
  di.allocate(0, count * 8);
  dout.allocate(0, actual.size() * 8);
  check(cudaMemcpy(dl.data, logs.data(), logs.size() * 4, cudaMemcpyHostToDevice));
  check(cudaMemcpy(dp.data, powers.data(), powers.size() * 4, cudaMemcpyHostToDevice));
  check(cudaMemcpy(di.data, indices.data(), count * 8, cudaMemcpyHostToDevice));
  positions<<<1, 32>>>(
      static_cast<uint64_t *>(di.data), static_cast<uint64_t *>(dout.data), count,
      omega, delta, static_cast<uint32_t *>(dl.data), static_cast<uint32_t *>(dp.data));
  check(cudaGetLastError());
  check(cudaDeviceSynchronize());
  check(
      cudaMemcpy(actual.data(), dout.data, actual.size() * 8, cudaMemcpyDeviceToHost));
  for (unsigned i = 0; i < count; i++)
    require(actual[3 * i] == expected[i] && actual[3 * i + 1] == expected[i] >> 32 &&
                actual[3 * i + 2] == (expected[i] & ((uint64_t(1) << 32) - 1)),
            "true34-bit canonical position/shard");
}
int main(int argc, char **argv) {
  try {
    require(argc == 2 || (argc == 3 && std::string(argv[2]) == "--small"),
            "fixture path required");
    bool small = argc == 3;
    Reader r(argv[1]);
    auto cases = r.scalar<uint32_t>();
    require(cases == 5, "fixture case count");
    unsigned passed = 0;
    for (unsigned i = 0; i < cases; i++) {
      Fixture f(r);
      for (unsigned count : {1u, 2u, 4u})
        for (int layout : {0, 1}) {
          if (small && (i >= 3 || count != 4))
            continue;
          verify<uint32_t>(f, count, layout);
          passed++;
          verify<uint64_t>(f, count, layout);
          passed++;
        }
      if (i == 0) {
        bool rejected = false;
        int id = 0;
        try {
          Resident<uint64_t> too_small(f.s, f.polynomial, f.delta, f.omega,
                                       f.logs.data(), f.logs.size(), f.powers.data(),
                                       f.powers.size(), &id, 1, f.n, 1);
        } catch (const std::exception &) {
          rejected = true;
        }
        require(rejected, "metadata cap accepted");
      }
    }
    wide_positions(r);
    require(r.file.get() == EOF, "trailing fixture bytes");
    std::cout << "{\"resident_code_cases\":" << passed << ",\"wide_positions\":true}"
              << std::endl;
    return 0;
  } catch (const std::exception &e) {
    std::cerr << e.what() << std::endl;
    return 1;
  }
}
