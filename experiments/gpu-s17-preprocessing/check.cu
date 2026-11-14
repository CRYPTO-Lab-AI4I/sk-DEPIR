// Bounded windows of the production s17 monomial expansion, including indices
// above2^32. The3^17 class cache fits; the full2GiB bitmap is NEVER allocated.
#include "preparation_kernels.cuh"
#include <iostream>
#include <vector>
using namespace liftrs_cuda::preparation;
void check(cudaError_t status) {
  if (status != cudaSuccess)
    throw std::runtime_error(cudaGetErrorString(status));
}
struct Device {
  void *pointer = nullptr;
  explicit Device(size_t bytes) { check(cudaMalloc(&pointer, bytes)); }
  ~Device() {
    if (pointer)
      cudaFree(pointer);
  }
  Device(const Device &) = delete;
  Device &operator=(const Device &) = delete;
};
// Four tiny physical buffers on GPU0 probe four LOGICAL slabs whose offsets
// exceed2^32. This checks production address arithmetic, not peer hardware.
__global__ void address_map(Shards<uint64_t> view, uint64_t *output) {
  const unsigned i = threadIdx.x;
  if (i < 16) {
    const size_t card = i / 4, offset = i % 4;
    output[i] = view[(card << 32) | offset];
  }
}
void check_address_map() {
  std::vector<std::unique_ptr<Device>> buffers;
  Shards<uint64_t> view;
  view.chunk = size_t(1) << 32;
  view.shift = 32;
  for (unsigned card = 0; card < 4; card++) {
    auto data = std::make_unique<Device>(32);
    uint64_t words[4];
    for (unsigned j = 0; j < 4; j++)
      words[j] = (uint64_t(card + 1) << 40) | j;
    check(cudaMemcpy(data->pointer, words, 32, cudaMemcpyHostToDevice));
    view.data[card] = static_cast<uint64_t *>(data->pointer);
    buffers.push_back(std::move(data));
  }
  Device result(128);
  address_map<<<1, 32>>>(view, static_cast<uint64_t *>(result.pointer));
  check(cudaGetLastError());
  check(cudaDeviceSynchronize());
  uint64_t output[16];
  check(cudaMemcpy(output, result.pointer, 128, cudaMemcpyDeviceToHost));
  for (unsigned i = 0; i < 16; i++)
    if (output[i] != ((uint64_t(i / 4 + 1) << 40) | (i % 4)))
      throw std::runtime_error("wide sharded address mapping");
  std::cout << "{\"kind\":\"address-map\",\"logical_points\":17179869184,"
               "\"physical_bytes\":256,\"match\":true}\n";
}
int main() {
  try {
    check(cudaSetDevice(0));
    check_address_map();
    constexpr unsigned s = 17, t = 5;
    constexpr uint64_t n = uint64_t(1) << 34;
    const size_t q = size_t(1) << s, words = 17;
    size_t classes = 1;
    for (unsigned j = 0; j < s; j++)
      classes *= 3;
    const size_t owned = classes + q * 4 + words * 8 + 8;
    if (owned > 256ull << 20)
      throw std::runtime_error("bounded window allocation cap");
    Device weights(classes), ternary(q * 4), bitmap(words * 8), counter(8);
    weight_classes<<<(classes + 255) / 256, 256>>>(
        static_cast<unsigned char *>(weights.pointer), classes, s, t);
    ternary_indices<<<(q + 255) / 256, 256>>>(
        static_cast<uint32_t *>(ternary.pointer), q, s);
    check(cudaGetLastError());
    check(cudaDeviceSynchronize());
    for (size_t first : {size_t(0), size_t(1) << 26, size_t(n / 64) - words}) {
      check(cudaMemset(counter.pointer, 0, 8));
      expand_good<<<1, 256>>>(
          static_cast<unsigned char *>(weights.pointer),
          static_cast<uint32_t *>(ternary.pointer),
          static_cast<uint64_t *>(bitmap.pointer),
          static_cast<unsigned long long *>(counter.pointer), first, words, n,
          s);
      check(cudaGetLastError());
      check(cudaDeviceSynchronize());
      std::vector<uint64_t> bits(words);
      uint64_t found = 0, expected_count = 0;
      check(cudaMemcpy(bits.data(), bitmap.pointer, words * 8,
                       cudaMemcpyDeviceToHost));
      check(cudaMemcpy(&found, counter.pointer, 8, cudaMemcpyDeviceToHost));
      for (size_t i = 0; i < words; i++)
        for (unsigned bit = 0; bit < 64; bit++) {
          const uint64_t index = 64 * (first + i) + bit;
          unsigned w[20];
          for (unsigned k = 0; k < s; k++)
            w[k] =
                unsigned((index >> k) & 1) + unsigned((index >> (s + k)) & 1);
          const bool expected = liftrs::preparation::good_weights(w, s, t);
          expected_count += expected;
          if (bool((bits[i] >> bit) & 1) != expected)
            throw std::runtime_error("wide GPU monomial window criterion");
        }
      if (found != expected_count)
        throw std::runtime_error("wide GPU monomial window count");
      std::cout << "{\"s\":17,\"first_index\":" << 64 * first
                << ",\"positions\":" << 64 * words
                << ",\"owned_device_bytes\":" << owned << ",\"match\":true}\n";
    }
    return 0;
  } catch (const std::exception &error) {
    std::cerr << error.what() << '\n';
    return 1;
  }
}
