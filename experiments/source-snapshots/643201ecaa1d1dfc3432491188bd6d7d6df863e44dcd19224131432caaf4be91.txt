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
size_t liftrs_plan_owned_bytes(const liftrs_plan* plan);
void liftrs_plan_destroy(liftrs_plan* plan);
#ifdef __cplusplus
}
#endif
#endif
