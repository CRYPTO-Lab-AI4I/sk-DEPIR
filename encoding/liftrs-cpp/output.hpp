#pragma once

// Output representations are separate from the shared field/scatter algorithm.
// Reused wider objects require atomic character writes even when logical output
// positions are distinct; do not access those objects through their old type
// concurrently. Packed output instead ORs disjoint bit fields in shared bytes.
#include <cstddef>
#include <cstdint>
#include <stdexcept>

namespace liftrs::output {
inline void store_native_u32(unsigned char *bytes, size_t point, uint32_t value) {
  const auto *source = reinterpret_cast<const unsigned char *>(&value);
  for (unsigned j = 0; j < 4; j++)
    __atomic_store_n(bytes + 4 * point + j, source[j], __ATOMIC_RELAXED);
}
inline void store_le_u16(unsigned char *bytes, size_t point, uint32_t value) {
  for (unsigned j = 0; j < 2; j++)
    __atomic_store_n(bytes + 2 * point + j,
                     static_cast<unsigned char>(value >> (8 * j)), __ATOMIC_RELAXED);
}
inline void or_packed(unsigned char *bytes, size_t point, uint32_t value,
                      unsigned width) {
  const size_t bit = point * width, first = bit / 8;
  const unsigned shift = bit % 8;
  const uint32_t bits = value << shift;
  for (unsigned j = 0; j < (shift + width + 7) / 8; j++)
    __atomic_fetch_or(bytes + first + j, static_cast<unsigned char>(bits >> (8 * j)),
                      __ATOMIC_RELAXED);
}
inline uint32_t load_le_u16(const unsigned char *bytes, size_t i) {
  return uint32_t(bytes[2 * i]) | (uint32_t(bytes[2 * i + 1]) << 8);
}
inline void narrow_inplace(uint32_t *data, size_t count, unsigned field_degree) {
  auto *bytes = reinterpret_cast<unsigned char *>(data);
  // Forward narrowing cannot overwrite a future source. The last coefficient
  // is unused by phi; keep its temporary representation zero.
  for (size_t i = 0; i < count; i++) {
    const uint32_t value = i < count - 1 ? data[i] : 0;
    if (value >= (uint32_t(1) << field_degree))
      throw std::runtime_error("product escaped base field before compaction");
    bytes[2 * i] = static_cast<unsigned char>(value);
    bytes[2 * i + 1] = static_cast<unsigned char>(value >> 8);
  }
}
inline void expand_inplace(uint32_t *data, size_t count) {
  const auto *bytes = reinterpret_cast<const unsigned char *>(data);
  // Reverse expansion leaves all not-yet-read packed values intact.
  for (size_t i = count; i-- > 0;)
    data[i] = load_le_u16(bytes, i);
}
} // namespace liftrs::output
