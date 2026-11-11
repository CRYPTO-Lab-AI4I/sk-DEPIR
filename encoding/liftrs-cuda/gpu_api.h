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
/* Opt-in limits; max_points <=2^32, device_cap includes all explicit allocations.
 * Legacy gm_create retains 2^22/512 MiB. Full domain support is a format bound,
 * not an allocation promise. Bounded creation requires payload plus one plan. */
int gm_create_bounded(unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t log_count,const uint32_t* powers,size_t power_count,size_t max_points,size_t device_cap,void** out,char*,size_t);
/* Experimental online-only two-work-array constructor. Same field/count/cap
 * bounds as bounded creation; requires2L<=N at SetOnline. Primitive Apply/Product
 * are rejected. Adds a small parity/error array; default constructors unchanged. */
int gm_create_reuse(unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t log_count,const uint32_t* powers,size_t power_count,size_t max_points,size_t device_cap,void** out,char*,size_t);
/* Explicit compact online experiment: direct-row transforms, one full arena,
 * bounded uint16 staging/assembly, half-sized bulk download and backwards host
 * expansion to owning uint32 output. s<=16 and2L<=N; no default switch. */
int gm_create_compact(unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t log_count,const uint32_t* powers,size_t power_count,size_t max_points,size_t device_cap,void** out,char*,size_t);
/* Last completed compact call: GPU compaction seconds, host expansion seconds,
 * bulk download bytes (excludes the error flag). Single-caller stats snapshot. */
int gm_compact_stats(void*,double* values,char*,size_t);
/* Transform-only experiment: direct_rows=0 owns data+transpose scratch;
 * direct_rows=1 owns data only and operates rows in their original layout.
 * Product and SetOnline are rejected. Default online/primitive modes unchanged. */
int gm_create_transform(unsigned direct_rows,unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t log_count,const uint32_t* powers,size_t power_count,size_t max_points,size_t device_cap,void** out,char*,size_t);
/* Descriptions: legacy [n,split,width,...] or v2 [0x324d474c,n_lo,n_hi,split,width,...].
 * Remaining rows/columns are unchanged. Check parses without device allocation;
 * Add also enforces engine limits and allocates workspace/plan. */
int gm_check_plan(void*,const uint32_t* words,size_t count,char*,size_t);
int gm_add_plan(void*,const uint32_t* words,size_t count,char*,size_t);
int gm_apply(void*,uint32_t* data,size_t n,int inverse,double* stats,char*,size_t);
int gm_convolution(void*,const uint32_t* a,const uint32_t* b,uint32_t* out,size_t n,double* stats,char*,size_t);
/* Unequal operands/low prefix. n must fit the FULL product, not just out_count.
 * Inputs are copied before output is written; unused device padding is cleared. */
int gm_product(void*,const uint32_t* a,size_t a_count,const uint32_t* b,size_t b_count,uint32_t* out,size_t out_count,size_t n,double* stats,char*,size_t);
/* Primary online encoding state. Code constants are copied once. Full spectrum
 * length must equal2^(2s); p_count=K+1. Require prepared full/prefix matrices.
 * stats[10] preserve first8 matrix slots and add buffer, canonical assembly. */
int gm_set_online(void*,const uint32_t* p,size_t p_count,const uint32_t* spectrum,size_t n,uint32_t omega,char*,size_t);
int gm_encode_online(void*,const uint32_t* message,size_t k,uint32_t* word,size_t n,double* stats,char*,size_t);
uint64_t gm_owned_bytes(void*);
void gm_destroy(void*);
#ifdef __cplusplus
}
#endif
