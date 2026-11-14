#ifndef LIFTRS_PREPARATION_API_H
#define LIFTRS_PREPARATION_API_H
#include <stddef.h>
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
/* Experimental single-caller public-code preparation. Width is selected from s;
 * base coefficients stay uint32. Limits precede all domain allocations. Handles
 * own native workspace/metadata and never retain borrowed host pointers. Byte
 * caps cover explicit payloads, not allocator/runtime or caller-owned arrays. */
typedef struct liftrs_preparation liftrs_preparation;
int liftrs_preparation_create(unsigned, uint32_t, uint32_t, uint64_t, unsigned, size_t,
                              uint64_t, liftrs_preparation **, char *, size_t);
int liftrs_preparation_good_count(liftrs_preparation *, unsigned, uint64_t *, char *,
                                  size_t);
int liftrs_preparation_good_fill(liftrs_preparation *, int64_t *, size_t, char *,
                                 size_t);
/* Direct bounded criterion witness checks; no domain/ternary-table allocation. */
int liftrs_preparation_good_range(unsigned, unsigned, uint64_t, unsigned char *, size_t,
                                  unsigned, char *, size_t);
int liftrs_preparation_apply(liftrs_preparation *, void *, size_t, unsigned, int,
                             char *, size_t);
int liftrs_preparation_grid(liftrs_preparation *, void *, size_t, unsigned, char *,
                            size_t);
int liftrs_preparation_product(liftrs_preparation *, const uint32_t *, size_t,
                               const uint32_t *, size_t, uint32_t *, size_t, size_t,
                               char *, size_t);
/* Each four-word factor is [constant,linear,quadratic,degree]; partner<i rows
 * are zero and ignored. Coordinates are sorted/nonnegative and exclude N-1. */
int liftrs_preparation_roots(liftrs_preparation *, const int64_t *, size_t, uint32_t *,
                             char *, size_t);
int liftrs_preparation_spectrum(liftrs_preparation *, const uint32_t *, size_t, void *,
                                size_t, unsigned, char *, size_t);
void liftrs_preparation_release_products(liftrs_preparation *);
uint64_t liftrs_preparation_bytes(const liftrs_preparation *);
void liftrs_preparation_destroy(liftrs_preparation *);
#ifdef __cplusplus
}
#endif
#endif
