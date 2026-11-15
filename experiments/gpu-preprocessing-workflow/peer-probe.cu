// Bounded diagnostic: CUDA peer copies first, then the SAME transpose kernel
// that stalled. No discovery, field arithmetic, Go runtime or protocol
// shortcut.
#include "preparation_kernels.cuh"
#include <chrono>
#include <ctime>
#include <iomanip>
#include <iostream>
#include <stdexcept>
#include <vector>
using namespace liftrs_cuda::preparation;
void check(cudaError_t e) {
  if (e != cudaSuccess)
    throw std::runtime_error(cudaGetErrorString(e));
}
int main() {
  constexpr unsigned cards = 4;
  constexpr size_t n = 65536, chunk = n / cards,
                   bytes = chunk * sizeof(uint32_t);
  const auto began = std::chrono::steady_clock::now();
  auto event = [&](const char *phase, const char *kind, size_t done,
                   size_t total) {
    const auto now =
        std::chrono::system_clock::to_time_t(std::chrono::system_clock::now());
    std::cout << "{\"utc\":\""
              << std::put_time(std::gmtime(&now), "%Y-%m-%dT%H:%M:%SZ")
              << "\",\"phase\":\"" << phase << "\",\"event\":\"" << kind
              << "\",\"completed\":" << done << ",\"total\":" << total
              << ",\"unit\":\"devices_or_links\",\"elapsed_seconds\":"
              << std::chrono::duration<double>(
                     std::chrono::steady_clock::now() - began)
                     .count()
              << "}" << std::endl;
  };
  try {
    int available = 0;
    check(cudaGetDeviceCount(&available));
    if (available < cards)
      throw std::runtime_error("four visible GPUs required");
    uint32_t *output[cards]{};
    Shards<uint32_t> src;
    src.chunk = chunk;
    src.shift = 14;
    std::vector<uint32_t> host(chunk);
    for (unsigned d = 0; d < cards; d++) {
      check(cudaSetDevice(d));
      check(cudaMalloc(&src.data[d], bytes));
      check(cudaMalloc(&output[d], bytes));
      for (size_t j = 0; j < chunk; j++)
        host[j] = uint32_t(d * chunk + j);
      check(
          cudaMemcpy(src.data[d], host.data(), bytes, cudaMemcpyHostToDevice));
      for (unsigned other = 0; other < cards; other++)
        if (other != d) {
          int allowed = 0;
          check(cudaDeviceCanAccessPeer(&allowed, d, other));
          if (!allowed)
            throw std::runtime_error("peer access unavailable");
          check(cudaDeviceEnablePeerAccess(other, 0));
        }
    }
    event("runtime-peer-copies", "started", 0, cards * (cards - 1));
    size_t done = 0;
    for (unsigned d = 0; d < cards; d++)
      for (unsigned other = 0; other < cards; other++)
        if (other != d) {
          check(cudaSetDevice(d));
          check(
              cudaMemcpyPeerAsync(output[d], d, src.data[other], other, bytes));
          check(cudaDeviceSynchronize());
          check(cudaMemcpy(host.data(), output[d], bytes,
                           cudaMemcpyDeviceToHost));
          for (size_t j = 0; j < chunk; j++)
            if (host[j] != other * chunk + j)
              throw std::runtime_error("peer copy mismatch");
          event("runtime-peer-copies", "progress", ++done, cards * (cards - 1));
        }
    event("runtime-peer-copies", "completed", done, done);
    event("peer-transpose-probe", "started", 0, cards);
    for (unsigned d = 0; d < cards; d++) {
      check(cudaSetDevice(d));
      transpose_tiles<<<16, dim3(32, 8)>>>(src, output[d], d * chunk, chunk,
                                           256, 256);
      check(cudaGetLastError());
    }
    for (unsigned d = 0; d < cards; d++) {
      check(cudaSetDevice(d));
      check(cudaDeviceSynchronize());
      check(cudaMemcpy(host.data(), output[d], bytes, cudaMemcpyDeviceToHost));
      for (size_t j = 0; j < chunk; j++) {
        const size_t i = d * chunk + j;
        if (host[j] != (i % 256) * 256 + i / 256)
          throw std::runtime_error("peer transpose mismatch");
      }
      event("peer-transpose-probe", "progress", d + 1, cards);
    }
    event("peer-transpose-probe", "completed", cards, cards);
    for (unsigned d = 0; d < cards; d++) {
      check(cudaSetDevice(d));
      check(cudaFree(output[d]));
      check(cudaFree(src.data[d]));
      for (unsigned other = 0; other < cards; other++)
        if (other != d)
          check(cudaDeviceDisablePeerAccess(other));
    }
    return 0;
  } catch (const std::exception &e) {
    event("peer-probe", "failed", 0, 1);
    std::cerr << e.what() << std::endl;
    return 1;
  }
}
