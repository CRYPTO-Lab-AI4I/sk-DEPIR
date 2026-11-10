#pragma once
#include <stddef.h>
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
/* Experimental, single-caller engine. Calls complete all borrowed host access
 * before returning. Inputs/descriptions are trusted canonical field values;
 * structural checks do not authenticate them. Device state is owned until destroy.
 * stats[8]: H2D, Taylor, columns, rows, layout, D2H, products, total seconds. */
int gm_create(unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t log_count,const uint32_t* powers,size_t power_count,void** out,char* error,size_t capacity);
int gm_add_plan(void*,const uint32_t* words,size_t count,char*,size_t);
int gm_apply(void*,uint32_t* data,size_t n,int inverse,double* stats,char*,size_t);
int gm_convolution(void*,const uint32_t* a,const uint32_t* b,uint32_t* out,size_t n,double* stats,char*,size_t);
/* Unequal operands/low prefix. n must fit the FULL product, not just out_count.
 * Inputs are copied before output is written; unused device padding is cleared. */
int gm_product(void*,const uint32_t* a,size_t a_count,const uint32_t* b,size_t b_count,uint32_t* out,size_t out_count,size_t n,double* stats,char*,size_t);
uint64_t gm_owned_bytes(void*);
void gm_destroy(void*);
#ifdef __cplusplus
}
#endif
