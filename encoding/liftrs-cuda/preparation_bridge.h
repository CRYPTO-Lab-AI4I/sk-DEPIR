// Dynamic loading keeps CUDA headers/runtime out of ordinary Go/CPU builds.
#ifndef LIFTRS_CUDA_PREPARATION_BRIDGE_H
#define LIFTRS_CUDA_PREPARATION_BRIDGE_H
#include "preparation_api.h"
#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>

typedef struct pg_bridge {
  void *library;
  void *engine;
  int (*apply)(void *, void *, size_t, unsigned, int, char *, size_t);
  int (*grid)(void *, void *, size_t, unsigned, char *, size_t);
  int (*product)(void *, const uint32_t *, size_t, const uint32_t *, size_t,
                 uint32_t *, size_t, size_t, char *, size_t);
  int (*spectrum)(void *, const uint32_t *, size_t, void *, size_t, unsigned,
                  char *, size_t);
  int (*spectrum_begin)(void *, const uint32_t *, size_t, size_t, char *, size_t);
  int (*spectrum_read)(void *, void *, size_t, size_t, unsigned, char *, size_t);
  int (*roots)(void *, const int64_t *, size_t, uint32_t *, char *, size_t);
  int (*good_count)(void *, unsigned, uint64_t *, char *, size_t);
  int (*good_fill)(void *, int64_t *, size_t, char *, size_t);
  int (*release)(void *, char *, size_t);
  int (*set_code)(void *, const uint32_t *, size_t, const void *, size_t, unsigned, int,
                  char *, size_t);
  int (*encode_code)(void *, const uint32_t *, size_t, uint32_t *, size_t, double *,
                     char *, size_t);
  int (*vector_stage)(void *, void *, size_t, unsigned, size_t, size_t, int, int,
                       size_t, size_t, char *, size_t);
  int (*taylor_degrees)(void *, unsigned, size_t, size_t, unsigned, uint64_t *, size_t, size_t *, char *, size_t);
  int (*taylor_stage)(void *, void *, size_t, unsigned, size_t, size_t, unsigned, int, size_t, size_t, size_t, char *, size_t);
  int (*stage_pointwise)(void *, void *, const void *, size_t, unsigned, size_t, size_t, char *, size_t);
  int (*rank_create)(void *, size_t, size_t, unsigned, unsigned, size_t, void **, char *, size_t);
  int (*rank_view)(void *, unsigned, size_t, size_t, int, char *, size_t);
  int (*rank_transfer)(void *, unsigned, size_t, void *, size_t, int, char *, size_t);
  int (*rank_local)(void *, unsigned, unsigned, char *, size_t);
  int (*rank_tile)(void *, unsigned, unsigned, size_t, void *, size_t, int, char *, size_t);
  int (*rank_vectors)(void *, unsigned, int, int, char *, size_t);
  int (*rank_degrees)(void *, unsigned, uint64_t *, size_t, size_t *, char *, size_t);
  int (*rank_taylor)(void *, unsigned, const void *, size_t, size_t, unsigned, int, size_t, size_t, size_t, char *, size_t);
  int (*rank_pointwise)(void *, unsigned, unsigned, char *, size_t);
  void (*rank_destroy)(void *);
  uint64_t (*owned)(void *);
  void (*destroy)(void *);
} pg_bridge;
static pg_bridge *pg_bridge_open(const char *path, unsigned s, uint32_t poly,
                                 uint32_t delta, uint64_t omega,
                                 const uint32_t *logs, size_t lc,
                                 const uint32_t *powers, size_t pc,
                                 const int *devices, size_t count,
                                 size_t maximum, uint64_t cap, char *error,
                                 size_t capacity) {
  pg_bridge *b = calloc(1, sizeof(pg_bridge));
  if (!b) {
    snprintf(error, capacity, "GPU preparation bridge allocation");
    return NULL;
  }
  b->library = dlopen(path, RTLD_NOW | RTLD_LOCAL);
  if (!b->library) {
    snprintf(error, capacity, "%s", dlerror());
    free(b);
    return NULL;
  }
  int (*create)(unsigned, uint32_t, uint32_t, uint64_t, const uint32_t *,
                size_t, const uint32_t *, size_t, const int *, size_t, size_t,
                uint64_t, void **, char *, size_t) =
      dlsym(b->library, "pg_create");
  b->apply = dlsym(b->library, "pg_apply");
  b->grid = dlsym(b->library, "pg_grid");
  b->product = dlsym(b->library, "pg_product");
  b->spectrum = dlsym(b->library, "pg_spectrum");
  b->spectrum_begin = dlsym(b->library, "pg_spectrum_begin");
  b->spectrum_read = dlsym(b->library, "pg_spectrum_read");
  b->roots = dlsym(b->library, "pg_roots");
  b->good_count = dlsym(b->library, "pg_good_count");
  b->good_fill = dlsym(b->library, "pg_good_fill");
  b->release = dlsym(b->library, "pg_release_products");
  // Optional additions: old preprocessing plugins remain compatible. Client
  // construction explicitly checks these before configuring any code data.
  b->set_code = dlsym(b->library, "pg_set_code");
  b->encode_code = dlsym(b->library, "pg_encode_code");
  b->vector_stage = dlsym(b->library, "pg_vector_stage");
  b->taylor_degrees = dlsym(b->library, "pg_taylor_degrees");
  b->taylor_stage = dlsym(b->library, "pg_taylor_stage");
  b->stage_pointwise = dlsym(b->library, "pg_stage_pointwise");
  b->rank_create=dlsym(b->library,"pg_rank_create");
  b->rank_view=dlsym(b->library,"pg_rank_view");
  b->rank_transfer=dlsym(b->library,"pg_rank_transfer");
  b->rank_local=dlsym(b->library,"pg_rank_local");
  b->rank_tile=dlsym(b->library,"pg_rank_tile");
  b->rank_vectors=dlsym(b->library,"pg_rank_vectors");
  b->rank_degrees=dlsym(b->library,"pg_rank_degrees");
  b->rank_taylor=dlsym(b->library,"pg_rank_taylor");
  b->rank_pointwise=dlsym(b->library,"pg_rank_pointwise");
  b->rank_destroy=dlsym(b->library,"pg_rank_destroy");
  b->owned = dlsym(b->library, "pg_owned_bytes");
  b->destroy = dlsym(b->library, "pg_destroy");
  if (!create || !b->apply || !b->grid || !b->product || !b->spectrum ||
      !b->roots || !b->good_count || !b->good_fill || !b->release ||
      !b->owned || !b->destroy) {
    snprintf(error, capacity, "resident GPU preprocessing ABI unavailable");
    dlclose(b->library);
    free(b);
    return NULL;
  }
  if (create(s, poly, delta, omega, logs, lc, powers, pc, devices, count,
             maximum, cap, &b->engine, error, capacity)) {
    dlclose(b->library);
    free(b);
    return NULL;
  }
  return b;
}
static int pg_bridge_has_rank(pg_bridge *b) {
 return b && b->rank_create && b->rank_view && b->rank_transfer && b->rank_local && b->rank_tile && b->rank_destroy;
}
static int pg_bridge_rank_create(pg_bridge *b,size_t rows,size_t columns,unsigned rank,unsigned ranks,size_t tile,void **out,char *error,size_t cap) {
 if(!pg_bridge_has_rank(b)){snprintf(error,cap,"GPU rank storage ABI unavailable");return 1;}
 return b->rank_create(b->engine,rows,columns,rank,ranks,tile,out,error,cap);
}
static int pg_bridge_rank_view(pg_bridge *b,void *h,unsigned slot,size_t rows,size_t columns,int interleaved,char *error,size_t cap) {
 return b->rank_view(h,slot,rows,columns,interleaved,error,cap);
}
static int pg_bridge_rank_transfer(pg_bridge *b,void *h,unsigned slot,size_t offset,void *values,size_t words,int upload,char *error,size_t cap) {
 return b->rank_transfer(h,slot,offset,values,words,upload,error,cap);
}
static int pg_bridge_rank_local(pg_bridge *b,void *h,unsigned source,unsigned target,char *error,size_t cap) {
 return b->rank_local(h,source,target,error,cap);
}
static int pg_bridge_rank_tile(pg_bridge *b,void *h,unsigned slot,unsigned peer,size_t offset,void *values,size_t words,int pack,char *error,size_t cap) {
 return b->rank_tile(h,slot,peer,offset,values,words,pack,error,cap);
}
static int pg_bridge_rank_vectors(pg_bridge *b,void *h,unsigned slot,int columns,int inverse,char *error,size_t cap) {
 if(!b->rank_vectors){snprintf(error,cap,"GPU resident vector ABI unavailable");return 1;}
 return b->rank_vectors(h,slot,columns,inverse,error,cap);
}
static int pg_bridge_rank_degrees(pg_bridge *b,void *h,unsigned level,uint64_t *out,size_t limit,size_t *count,char *error,size_t cap) {
 if(!b->rank_degrees){snprintf(error,cap,"GPU resident Taylor ABI unavailable");return 1;}
 return b->rank_degrees(h,level,out,limit,count,error,cap);
}
static int pg_bridge_rank_taylor(pg_bridge *b,void *h,unsigned slot,const void *snapshot,size_t words,size_t capacity,unsigned level,int quotient,size_t block,size_t first,size_t count,char *error,size_t cap) {
 if(!b->rank_taylor){snprintf(error,cap,"GPU resident Taylor ABI unavailable");return 1;}
 return b->rank_taylor(h,slot,snapshot,words,capacity,level,quotient,block,first,count,error,cap);
}
static int pg_bridge_rank_pointwise(pg_bridge *b,void *h,unsigned left,unsigned right,char *error,size_t cap) {
 if(!b->rank_pointwise){snprintf(error,cap,"GPU resident pointwise ABI unavailable");return 1;}
 return b->rank_pointwise(h,left,right,error,cap);
}
static void pg_bridge_rank_destroy(pg_bridge *b,void *h){b->rank_destroy(h);}
static int pg_bridge_has_vector_stage(pg_bridge *b) {
  return b && b->vector_stage;
}
static int pg_bridge_vector_stage(pg_bridge *b, void *data, size_t words,
    unsigned width, size_t domain, size_t local_capacity, int columns,
    int inverse, size_t first, size_t count, char *error, size_t cap) {
  if (!pg_bridge_has_vector_stage(b)) {
    snprintf(error, cap, "GPU vector stage ABI unavailable");
    return 1;
  }
  return b->vector_stage(b->engine, data, words, width, domain, local_capacity,
                         columns, inverse, first, count, error, cap);
}
static int pg_bridge_taylor_degrees(pg_bridge *b, unsigned width, size_t domain,
    size_t local_capacity, unsigned level, uint64_t *out, size_t limit, size_t *count,
    char *error, size_t cap) {
  if (!b || !b->taylor_degrees || !b->taylor_stage) {
    snprintf(error, cap, "GPU Taylor stage ABI unavailable"); return 1;
  }
  return b->taylor_degrees(b->engine,width,domain,local_capacity,level,out,limit,count,error,cap);
}
static int pg_bridge_taylor_stage(pg_bridge *b, void *data, size_t words, unsigned width,
    size_t domain, size_t local_capacity, unsigned level, int quotient, size_t block,
    size_t first, size_t count, char *error, size_t cap) {
  if (!b || !b->taylor_stage) {snprintf(error, cap, "GPU Taylor stage ABI unavailable");return 1;}
  return b->taylor_stage(b->engine,data,words,width,domain,local_capacity,level,quotient,block,first,count,error,cap);
}
static int pg_bridge_stage_pointwise(pg_bridge *b, void *left, const void *right,
    size_t words, unsigned width, size_t domain, size_t local_capacity,
    char *error, size_t cap) {
  if (!b || !b->stage_pointwise) {snprintf(error,cap,"GPU stage product ABI unavailable");return 1;}
  return b->stage_pointwise(b->engine,left,right,words,width,domain,local_capacity,error,cap);
}
static int pg_bridge_apply(pg_bridge *b, void *values, size_t n, unsigned width,
                           int inverse, char *error, size_t cap) {
  return b->apply(b->engine, values, n, width, inverse, error, cap);
}
static int pg_bridge_grid(pg_bridge *b, void *values, size_t n, unsigned width,
                          char *error, size_t cap) {
  return b->grid(b->engine, values, n, width, error, cap);
}
static int pg_bridge_product(pg_bridge *b, const uint32_t *a, size_t ac,
                             const uint32_t *c, size_t cc, uint32_t *out,
                             size_t count, size_t n, char *error, size_t cap) {
  return b->product(b->engine, a, ac, c, cc, out, count, n, error, cap);
}
static int pg_bridge_spectrum(pg_bridge *b, const uint32_t *g, size_t count,
                              void *out, size_t n, unsigned width, char *error,
                              size_t cap) {
  return b->spectrum(b->engine, g, count, out, n, width, error, cap);
}
static int pg_bridge_spectrum_begin(pg_bridge *b, const uint32_t *g, size_t count,
                                     size_t n, char *error, size_t cap) {
  if (!b->spectrum_begin || !b->spectrum_read) {
    snprintf(error, cap, "GPU streamed spectrum ABI unavailable");
    return 1;
  }
  return b->spectrum_begin(b->engine, g, count, n, error, cap);
}
static int pg_bridge_spectrum_read(pg_bridge *b, void *out, size_t first,
                                    size_t count, unsigned width,
                                    char *error, size_t cap) {
  if (!b->spectrum_read) {
    snprintf(error, cap, "GPU streamed spectrum ABI unavailable");
    return 1;
  }
  return b->spectrum_read(b->engine, out, first, count, width, error, cap);
}
static int pg_bridge_roots(pg_bridge *b, const int64_t *indices, size_t count,
                           uint32_t *output, char *error, size_t cap) {
  return b->roots(b->engine, indices, count, output, error, cap);
}
static int pg_bridge_good_count(pg_bridge *b, unsigned t, uint64_t *count,
                                char *error, size_t cap) {
  return b->good_count(b->engine, t, count, error, cap);
}
static int pg_bridge_good_fill(pg_bridge *b, int64_t *values, size_t count,
                               char *error, size_t cap) {
  return b->good_fill(b->engine, values, count, error, cap);
}
static int pg_bridge_release(pg_bridge *b, char *error, size_t capacity) {
  return b->release(b->engine, error, capacity);
}
static int pg_bridge_set_code(pg_bridge *b, const uint32_t *p, size_t pc,
                              const void *spectrum, size_t n, unsigned width,
                              int layout, char *error, size_t cap) {
  if (!b || !b->set_code || !b->encode_code) {
    snprintf(error, cap, "resident GPU client encoding ABI unavailable");
    return 1;
  }
  return b->set_code(b->engine, p, pc, spectrum, n, width, layout, error, cap);
}
static int pg_bridge_encode_code(pg_bridge *b, const uint32_t *message, size_t k,
                                 uint32_t *word, size_t n, double *stats, char *error,
                                 size_t cap) {
  if (!b || !b->encode_code) {
    snprintf(error, cap, "resident GPU client encoding ABI unavailable");
    return 1;
  }
  return b->encode_code(b->engine, message, k, word, n, stats, error, cap);
}
static uint64_t pg_bridge_owned(pg_bridge *b) {
  return b ? b->owned(b->engine) : 0;
}
static void pg_bridge_close(pg_bridge *b) {
  if (b) {
    b->destroy(b->engine);
    dlclose(b->library);
    free(b);
  }
}
#endif
