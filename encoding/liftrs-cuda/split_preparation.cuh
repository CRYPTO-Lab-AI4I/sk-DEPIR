#pragma once
#include <limits>

// Explicit block convolution over the existing resident Product. Include after
// preparation.cu. Host admission covers output plus one temporary, excluding
// caller inputs/references, allocator/runtime and filesystem overhead.
namespace {
using ProductClock = std::chrono::steady_clock;
inline double product_elapsed(ProductClock::time_point start) {
  return std::chrono::duration<double>(ProductClock::now() - start).count();
}
struct ProductStats {
  double seconds = 0, assembly_seconds = 0;
  size_t calls = 0, largest_domain = 0;
  uint64_t h2d_coeff_bytes = 0, d2h_coeff_bytes = 0, temporary_host_bytes = 0;
};
void progress(const std::string &phase, const char *event, size_t done,
              size_t total, ProductClock::time_point began) {
  liftrs::execution::progress_record(
      phase.c_str(), "preprocessing-components", event, done, total,
      uint64_t(product_elapsed(began) * 1000), "subproducts");
}
size_t product_domain(size_t coefficients) {
  size_t n = 4;
  while (n < coefficients) {
    if (n > std::numeric_limits<size_t>::max() / 2)
      throw std::runtime_error("product domain overflow");
    n *= 2;
  }
  return n;
}
template <class Word>
std::vector<uint32_t> product(Resident<Word> &engine,
                              const std::vector<uint32_t> &a,
                              const std::vector<uint32_t> &b, size_t block,
                              uint64_t host_cap, ProductStats &stats) {
  if (a.empty() || b.empty() ||
      a.size() > std::numeric_limits<size_t>::max() - b.size() + 1 ||
      (block && (block < 4 || block > std::numeric_limits<size_t>::max() / 2)))
    throw std::runtime_error("split product input/block");
  const size_t output_size = a.size() + b.size() - 1;
  const size_t temporary = block ? std::min(output_size, 2 * block - 1) : 0;
  if (output_size > host_cap / 4 || temporary > host_cap / 4 - output_size)
    throw std::runtime_error("split product host output/temp admission");
  stats = {};
  const auto began = ProductClock::now();
  auto last = began;
  std::vector<uint32_t> out(output_size);
  if (!block) {
    const size_t n = product_domain(output_size);
    engine.product(a.data(), a.size(), b.data(), b.size(), out.data(), out.size(), n);
    stats.calls = 1;
    stats.largest_domain = n;
    stats.h2d_coeff_bytes = 4 * (a.size() + b.size());
    stats.d2h_coeff_bytes = 4 * out.size();
  } else {
    const size_t count = ((a.size() + block - 1) / block) *
                         ((b.size() + block - 1) / block);
    progress("split-products", "started", 0, count, began);
    for (size_t first = 0; first < a.size(); first += block) {
      for (size_t second = 0; second < b.size(); second += block) {
        const size_t ac = std::min(block, a.size() - first);
        const size_t bc = std::min(block, b.size() - second);
        const size_t take = ac + bc - 1, n = product_domain(take);
        std::vector<uint32_t> temporary(take);
        engine.product(a.data() + first, ac, b.data() + second, bc,
                       temporary.data(), take, n);
        const auto packing = ProductClock::now();
        // Polynomial addition is XOR in characteristic two. Inputs stay immutable;
        // only the output is accumulated, including overlaps between block pairs.
        for (size_t j = 0; j < take; j++)
          out[first + second + j] ^= temporary[j];
        stats.assembly_seconds += product_elapsed(packing);
        stats.calls++;
        stats.largest_domain = std::max(stats.largest_domain, n);
        stats.temporary_host_bytes = std::max(stats.temporary_host_bytes, uint64_t(4 * take));
        stats.h2d_coeff_bytes += 4 * (ac + bc);
        stats.d2h_coeff_bytes += 4 * take;
        if (ProductClock::now() - last >= std::chrono::seconds(30)) {
          progress("split-products", "progress", stats.calls, count, began);
          last = ProductClock::now();
        }
      }
    }
    progress("split-products", "completed", stats.calls, count, began);
  }
  stats.seconds = product_elapsed(began);
  return out;
}
} // namespace
