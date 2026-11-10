//go:build liftrs_cuda && cgo

// Package gpumatrix is an experiment-only dynamic bridge to the CUDA matrix
// prototype. It adds no CUDA dependency to ordinary builds or library APIs.
package gpumatrix

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>
#include "gpu_api.h"
typedef struct {
 void *library,*engine;
 int (*create)(unsigned,unsigned,unsigned,const uint32_t*,size_t,const uint32_t*,size_t,void**,char*,size_t);
 int (*add)(void*,const uint32_t*,size_t,char*,size_t);
 int (*apply)(void*,uint32_t*,size_t,int,double*,char*,size_t);
 uint64_t (*owned)(void*);
 void (*destroy)(void*);
} bridge;
static bridge* bridge_open(const char* path,unsigned s,unsigned poly,unsigned delta,const uint32_t* logs,size_t lc,const uint32_t* powers,size_t pc,char* error,size_t capacity){
 bridge* b=calloc(1,sizeof(bridge));if(!b)return NULL;
 b->library=dlopen(path,RTLD_NOW|RTLD_LOCAL);
 if(!b->library){snprintf(error,capacity,"%s",dlerror());free(b);return NULL;}
 b->create=dlsym(b->library,"gm_create");b->add=dlsym(b->library,"gm_add_plan");b->apply=dlsym(b->library,"gm_apply");b->owned=dlsym(b->library,"gm_owned_bytes");b->destroy=dlsym(b->library,"gm_destroy");
 if(!b->create||!b->add||!b->apply||!b->owned||!b->destroy){snprintf(error,capacity,"missing GPU ABI symbol");dlclose(b->library);free(b);return NULL;}
 if(b->create(s,poly,delta,logs,lc,powers,pc,&b->engine,error,capacity)){dlclose(b->library);free(b);return NULL;}return b;
}
static int bridge_add(bridge* b,const uint32_t* words,size_t count,char* error,size_t capacity){return b->add(b->engine,words,count,error,capacity);}
static int bridge_apply(bridge* b,uint32_t* words,size_t count,int inverse,double* stats,char* error,size_t capacity){return b->apply(b->engine,words,count,inverse,stats,error,capacity);}
static uint64_t bridge_owned(bridge* b){return b->owned(b->engine);}
static void bridge_close(bridge* b){if(b){b->destroy(b->engine);dlclose(b->library);free(b);}}
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Engine serializes all calls and closes owned CUDA state explicitly. Host
// slices are borrowed only until synchronous Apply returns, never retained.
type Engine struct {
	mu     sync.Mutex
	handle *C.bridge
}

// New loads the explicitly supplied experimental shared library and copies
// trusted base-field tables. The caller must use a checksum-verified binary.
func New(path string, s int, polynomial, delta uint32, logs, powers []uint32) (*Engine, error) {
	if s < 2 || s > 16 || len(logs) != 1<<uint(s) || len(powers) != 2*((1<<uint(s))-1) {
		return nil, fmt.Errorf("GPU table shape")
	}
	name := C.CString(path)
	defer C.free(unsafe.Pointer(name))
	var message [512]C.char
	h := C.bridge_open(name, C.uint(s), C.uint(polynomial), C.uint(delta), (*C.uint32_t)(unsafe.Pointer(&logs[0])), C.size_t(len(logs)), (*C.uint32_t)(unsafe.Pointer(&powers[0])), C.size_t(len(powers)), &message[0], 512)
	runtime.KeepAlive(logs)
	runtime.KeepAlive(powers)
	if h == nil {
		return nil, fmt.Errorf("GPU initialization: %s", C.GoString(&message[0]))
	}
	return &Engine{handle: h}, nil
}

// AddPlan copies a bounded trusted, pointer-free matrix description.
func (e *Engine) AddPlan(words []uint32) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.handle == nil || len(words) == 0 {
		return fmt.Errorf("GPU closed/empty plan")
	}
	var message [512]C.char
	code := C.bridge_add(e.handle, (*C.uint32_t)(unsafe.Pointer(&words[0])), C.size_t(len(words)), &message[0], 512)
	runtime.KeepAlive(words)
	if code != 0 {
		return fmt.Errorf("GPU plan: %s", C.GoString(&message[0]))
	}
	return nil
}

// Apply transforms canonical packed coefficients in place. It includes host/
// device copies; the returned stage seconds use the experimental ABI ordering.
func (e *Engine) Apply(data []uint32, inverse bool) ([8]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var stats [8]C.double
	var out [8]float64
	if e.handle == nil || len(data) == 0 {
		return out, fmt.Errorf("GPU closed/empty buffer")
	}
	var message [512]C.char
	flag := 0
	if inverse {
		flag = 1
	}
	code := C.bridge_apply(e.handle, (*C.uint32_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), C.int(flag), &stats[0], &message[0], 512)
	runtime.KeepAlive(data)
	for i := range out {
		out[i] = float64(stats[i])
	}
	if code != 0 {
		return out, fmt.Errorf("GPU apply: %s", C.GoString(&message[0]))
	}
	return out, nil
}

// OwnedBytes counts explicit device tables/plans/buffers, excluding driver/context.
func (e *Engine) OwnedBytes() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.handle == nil {
		return 0
	}
	return uint64(C.bridge_owned(e.handle))
}

// Close releases plans, buffers and the dynamic library after active calls finish.
func (e *Engine) Close() { e.mu.Lock(); defer e.mu.Unlock(); C.bridge_close(e.handle); e.handle = nil }
