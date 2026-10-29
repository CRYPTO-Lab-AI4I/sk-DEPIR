#ifndef LIFTRS_C_API_H
#define LIFTRS_C_API_H
#include <stddef.h>
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
/* Immutable C++-owned plan. Create copies the bounded, pointer-free Go plan
 * description. Apply borrows an existing uint32 buffer only until return; there
 * is no per-element boundary call, retained Go pointer, or input/output copy. */
typedef struct liftrs_plan liftrs_plan;
int liftrs_plan_create(const uint32_t* words, size_t count, liftrs_plan** out,
                       char* error, size_t error_capacity);
int liftrs_plan_apply(const liftrs_plan* plan, uint32_t* data, size_t count,
                      int inverse, uint64_t* kernel_ns,
                      char* error, size_t error_capacity);
/* General-basis affine execution; start is a packed field point, not a row index. */
int liftrs_plan_apply_at(const liftrs_plan*, uint32_t*, size_t, uint32_t start,
                        int inverse, uint64_t*, char*, size_t);
/* In-memory matrix decomposition. scratch must be disjoint and count elements.
 * rows includes S_split; columns uses S_split(high basis). Both are immutable. */
int liftrs_matrix_apply(const liftrs_plan* rows, const liftrs_plan* columns,
                        uint32_t* data, uint32_t* scratch, size_t count,
                        unsigned split, unsigned workers, int inverse, uint64_t* stage_ns,
                        char* error, size_t error_capacity);
/* One-time code-only preparation; caller must exclude concurrent execution. */
int liftrs_plan_prepare_taylor(liftrs_plan*, unsigned high_levels, char*, size_t);
size_t liftrs_plan_owned_bytes(const liftrs_plan* plan);
void liftrs_plan_destroy(liftrs_plan* plan);
#ifdef __cplusplus
}
#endif
#endif
