// Count only: use the production membership criterion and exact class sizes.
// This does not build/certify sampled support or admit s18 transform execution.
#include "../../encoding/liftrs-cuda/preparation_kernels.cuh"
#include <chrono>
#include <cstdint>
#include <ctime>
#include <cuda_runtime.h>
#include <iostream>
#include <stdexcept>
#include <string>
#include <vector>

void check(cudaError_t status) {
  if (status != cudaSuccess)
    throw std::runtime_error(cudaGetErrorString(status));
}
__global__ void weighted_dimension(unsigned s, unsigned t, uint64_t first,
                                   uint64_t count, unsigned long long *total) {
  uint64_t i = uint64_t(blockIdx.x) * blockDim.x + threadIdx.x;
  unsigned long long value = 0;
  if (i < count) {
    unsigned ones = 0;
    if (liftrs::preparation::good_weight_class(first + i, s, t, &ones))
      value = 1ull << ones;
  }
  __shared__ unsigned long long sums[256];
  sums[threadIdx.x] = value;
  __syncthreads();
  for (unsigned step = 128; step; step >>= 1) {
    if (threadIdx.x < step)
      sums[threadIdx.x] += sums[threadIdx.x + step];
    __syncthreads();
  }
  if (threadIdx.x == 0)
    atomicAdd(total, sums[0]);
}
struct Counter {
  unsigned long long *value = nullptr;
  Counter() {
    check(cudaMalloc(&value, 8));
  }
  ~Counter() {
    if (value)
      cudaFree(value);
  }
};
void event(const char *kind, unsigned s, unsigned t, uint64_t completed, uint64_t total,
           double seconds, uint64_t k = 0) {
  const auto now = std::time(nullptr);
  std::tm utc{};
  gmtime_r(&now, &utc);
  char text[32];
  std::strftime(text, sizeof(text), "%Y-%m-%dT%H:%M:%SZ", &utc);
  std::cout << "{\"utc\":\"" << text
            << "\",\"phase\":\"exact-weighted-dimension\",\"event\":\"" << kind
            << "\",\"s\":" << s << ",\"t\":" << t << ",\"completed\":" << completed
            << ",\"total\":" << total
            << ",\"unit\":\"weight classes\",\"seconds\":" << seconds << ",\"k\":" << k
            << ",\"n\":" << (1ull << (2 * s)) << ",\"owned_device_bytes\":8}"
            << std::endl;
}
int main(int argc, char **argv) {
  try {
    if (argc != 3)
      throw std::runtime_error("arguments: s t");
    auto s = std::stoul(argv[1]), t = std::stoul(argv[2]);
    if (s < 2 || s > 18 || !t || t >= (1ull << s) - 1)
      throw std::runtime_error("bounded dimension parameters");
    check(cudaSetDevice(0));
    Counter total;
    check(cudaMemset(total.value, 0, 8));
    uint64_t classes = 1;
    for (unsigned j = 0; j < s; j++)
      classes *= 3;
    const auto began = std::chrono::steady_clock::now();
    auto last = began;
    event("started", s, t, 0, classes, 0);
    constexpr uint64_t wave = 1 << 20;
    for (uint64_t first = 0; first < classes; first += wave) {
      const uint64_t count = std::min(wave, classes - first);
      weighted_dimension<<<(count + 255) / 256, 256>>>(s, t, first, count, total.value);
      check(cudaGetLastError());
      check(cudaDeviceSynchronize());
      auto now = std::chrono::steady_clock::now();
      if (now - last >= std::chrono::seconds(30)) {
        event("progress", s, t, first + count, classes,
              std::chrono::duration<double>(now - began).count());
        last = now;
      }
    }
    uint64_t dimension = 0;
    check(cudaMemcpy(&dimension, total.value, 8, cudaMemcpyDeviceToHost));
    if (!dimension || dimension >= (1ull << (2 * s)))
      throw std::runtime_error("invalid dimension");
    event(
        "completed", s, t, classes, classes,
        std::chrono::duration<double>(std::chrono::steady_clock::now() - began).count(),
        dimension);
    return 0;
  } catch (const std::exception &e) {
    std::cerr << e.what() << std::endl;
    return 1;
  }
}
