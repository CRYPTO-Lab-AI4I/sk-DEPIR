#ifndef LIFTRS_CUDA_PREPARATION_API_H
#define LIFTRS_CUDA_PREPARATION_API_H
#include <stddef.h>
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
// Explicit same-node peer engine; caps are per device and exclude CUDA context.
// Creation builds metadata/tables only. No borrowed host pointer is retained.
int pg_create(unsigned s, uint32_t polynomial, uint32_t delta, uint64_t omega,
              const uint32_t *logs, size_t log_count, const uint32_t *powers,
              size_t power_count, const int *devices, size_t device_count,
              size_t maximum, uint64_t per_device_cap, void **output,
              char *error, size_t capacity);
int pg_apply(void *, void *, size_t, unsigned, int, char *, size_t);
int pg_grid(void *, void *, size_t, unsigned, char *, size_t);
int pg_product(void *, const uint32_t *, size_t, const uint32_t *, size_t,
               uint32_t *, size_t, size_t, char *, size_t);
int pg_spectrum(void *, const uint32_t *, size_t, void *, size_t, unsigned,
                char *, size_t);
// Optional bounded-host export: begin computes one zero-padded forward FFT;
// read copies canonical contiguous slots. Serialize through the final read.
// Any subsequent operation that grows/reuses transform storage invalidates it.
int pg_spectrum_begin(void *, const uint32_t *, size_t, size_t, char *, size_t);
int pg_spectrum_read(void *, void *, size_t, size_t, unsigned, char *, size_t);
int pg_roots(void *, const int64_t *, size_t, uint32_t *, char *, size_t);
int pg_good_count(void *, unsigned, uint64_t *, char *, size_t);
int pg_good_fill(void *, int64_t *, size_t, char *, size_t);
int pg_release_products(void *, char *, size_t);
// Optional resident client ABI. Constants are copied once; no host pointer is
// retained. layout0=peer,1=local; width describes spectrum slots (4/8).
int pg_set_code(void *, const uint32_t *, size_t, const void *, size_t, unsigned, int,
                char *, size_t);
int pg_encode_code(void *, const uint32_t *, size_t, uint32_t *, size_t, double *,
                   char *, size_t);
// Optional complete-vector stage: interleaved local tile, immutable global plan
// and capacity. Use a dedicated engine; do not mix resident operations afterward.
int pg_vector_stage(void *, void *, size_t, unsigned, size_t, size_t, int, int,
                    size_t, size_t, char *, size_t);
uint64_t pg_owned_bytes(void *);
void pg_destroy(void *);
#ifdef __cplusplus
}
#endif
#endif
