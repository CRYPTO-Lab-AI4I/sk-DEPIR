#ifndef LIFTRS_WIDE_API_H
#define LIFTRS_WIDE_API_H
#include <stddef.h>
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
/* Separate experimental64-bit handles. Field parameters are trusted code data;
 * constructors validate shape/ranges, not code authenticity. Plans copy their
 * basis and retain no caller pointer. Apply borrows canonical buffers only for
 * the call; matrix data/scratch must be wholly disjoint. No old ABI is widened. */
typedef struct liftrs_wide_plan liftrs_wide_plan;
typedef struct liftrs_wide_matrix liftrs_wide_matrix;
int liftrs_wide_plan_create(unsigned,uint32_t,uint32_t,const uint64_t*,size_t,liftrs_wide_plan**,char*,size_t);
int liftrs_wide_plan_apply(const liftrs_wide_plan*,uint64_t*,size_t,uint64_t,int,uint64_t*,char*,size_t);
size_t liftrs_wide_plan_bytes(const liftrs_wide_plan*);
void liftrs_wide_plan_destroy(liftrs_wide_plan*);
/* Canonical binary grid, dim<=2s,1<=split<dim. Immutable after construction.
 * native_ns includes canonical-input validation; stages[3] are Taylor/columns/
 * rows only. Apply is in-place and errors may leave private buffers partial. */
int liftrs_wide_matrix_create(unsigned,uint32_t,uint32_t,unsigned,unsigned,liftrs_wide_matrix**,char*,size_t);
int liftrs_wide_matrix_apply(const liftrs_wide_matrix*,uint64_t*,uint64_t*,size_t,unsigned,int,uint64_t*,uint64_t*,char*,size_t);
size_t liftrs_wide_matrix_bytes(const liftrs_wide_matrix*);
void liftrs_wide_matrix_destroy(liftrs_wide_matrix*);
/* Complete experimental cyclic schedule. Code/message buffers are read-only
 * borrowed for one call. Exactly one spectrum pointer is supplied; all
 * values must fit the field. The shared loader selects canonical slot widths. arena=max(2L,N)64-bit values, h=K32-bit values,
 * scratch=N64-bit values, all mutually disjoint and disjoint from input/code.
 * Output is written as N native uint32 values into scratch's first4N bytes using
 * byte stores; its retained backing remains8N. No caller pointer is retained.
 * stats[7]: full native ns,Taylor,columns,rows,buffer/validation,pointwise,assembly.
 * Construction checks primitive omega before parallel scatter; field/code data
 * authenticity and code identities remain the caller's responsibility. */
typedef struct liftrs_wide_online liftrs_wide_online;
int liftrs_wide_online_create(unsigned,uint32_t,uint32_t,uint64_t,uint64_t,liftrs_wide_online**,char*,size_t);
int liftrs_wide_online_encode(const liftrs_wide_online*,unsigned,const uint32_t*,const uint32_t*,const uint32_t*,const uint64_t*,uint64_t*,size_t,uint32_t*,size_t,uint64_t*,size_t,uint64_t*,char*,size_t);
size_t liftrs_wide_online_bytes(const liftrs_wide_online*);
void liftrs_wide_online_destroy(liftrs_wide_online*);
#ifdef __cplusplus
}
#endif
#endif
