//go:build liftrs_cuda && cgo

// Package liftrscuda provides the optional dynamically loaded CUDA backend.
// Ordinary CPU builds require no CUDA headers, toolkit or runtime.
package liftrscuda

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>
#include "gpu_api.h"
typedef struct {
 void *library,*engine;
 int (*create)(unsigned,unsigned,unsigned,const uint32_t*,size_t,const uint32_t*,size_t,void**,char*,size_t);
 int (*bounded)(unsigned,unsigned,unsigned,const uint32_t*,size_t,const uint32_t*,size_t,size_t,size_t,void**,char*,size_t);
 int (*transform)(unsigned,unsigned,unsigned,unsigned,const uint32_t*,size_t,const uint32_t*,size_t,size_t,size_t,void**,char*,size_t);
 int (*check)(void*,const uint32_t*,size_t,char*,size_t);
 int (*add)(void*,const uint32_t*,size_t,char*,size_t);
 int (*apply)(void*,uint32_t*,size_t,int,double*,char*,size_t);
 int (*product)(void*,const uint32_t*,size_t,const uint32_t*,size_t,uint32_t*,size_t,size_t,double*,char*,size_t);
 int (*set_online)(void*,const uint32_t*,size_t,const uint32_t*,size_t,uint32_t,char*,size_t);
 int (*encode_online)(void*,const uint32_t*,size_t,uint32_t*,size_t,double*,char*,size_t);
 int (*encode_streamed)(void*,const uint32_t*,size_t,const uint32_t*,size_t,const uint32_t*,size_t,uint32_t*,size_t,size_t,double*,char*,size_t);
 int (*stream_stats)(void*,double*,char*,size_t);
 int (*prepare_packed)(void*,const uint32_t*,size_t,const uint32_t*,size_t,const uint32_t*,size_t,size_t,size_t,double*,char*,size_t);
 int (*read_packed)(void*,size_t,unsigned char*,size_t,double*,char*,size_t);
 int (*clear_packed)(void*,char*,size_t);
 int (*encode_backed)(void*,const uint32_t*,size_t,const uint32_t*,const uint32_t*,size_t,uint32_t*,uint32_t*,size_t,uint32_t,double*,char*,size_t);
 int (*apply_backed)(void*,uint32_t*,size_t,int,double*,char*,size_t);
 int (*compact_stats)(void*,double*,char*,size_t);
 uint64_t (*owned)(void*);
 void (*destroy)(void*);
} bridge;
static bridge* bridge_open(const char* path,unsigned s,unsigned poly,unsigned delta,const uint32_t* logs,size_t lc,const uint32_t* powers,size_t pc,size_t max_points,size_t device_cap,size_t tile,int bounded,char* error,size_t capacity){
 bridge* b=calloc(1,sizeof(bridge));if(!b)return NULL;
 b->library=dlopen(path,RTLD_NOW|RTLD_LOCAL);
 if(!b->library){snprintf(error,capacity,"%s",dlerror());free(b);return NULL;}
 b->create=dlsym(b->library,"gm_create");b->add=dlsym(b->library,"gm_add_plan");b->apply=dlsym(b->library,"gm_apply");b->owned=dlsym(b->library,"gm_owned_bytes");b->destroy=dlsym(b->library,"gm_destroy");
 b->bounded=dlsym(b->library,bounded==2?"gm_create_reuse":bounded==5?"gm_create_compact":bounded==6?"gm_create_streamed":"gm_create_bounded");b->check=dlsym(b->library,"gm_check_plan");
 b->set_online=dlsym(b->library,"gm_set_online");b->encode_online=dlsym(b->library,"gm_encode_online");
 b->encode_streamed=dlsym(b->library,"gm_encode_streamed");b->stream_stats=dlsym(b->library,"gm_stream_stats");
 b->prepare_packed=dlsym(b->library,"gm_prepare_packed");b->read_packed=dlsym(b->library,"gm_read_packed");b->clear_packed=dlsym(b->library,"gm_clear_packed");
 b->compact_stats=dlsym(b->library,"gm_compact_stats");
 b->transform=dlsym(b->library,bounded==7?"gm_create_backed":"gm_create_transform");b->apply_backed=dlsym(b->library,"gm_apply_backed");
 b->encode_backed=dlsym(b->library,"gm_encode_backed");
 b->product=dlsym(b->library,"gm_product"); // Optional for old transform-only libraries.
 if(!b->create||!b->add||!b->apply||!b->owned||!b->destroy){snprintf(error,capacity,"missing GPU ABI symbol");dlclose(b->library);free(b);return NULL;}
 if((bounded==3||bounded==4||bounded==7)&&!b->transform){snprintf(error,capacity,"transform layout ABI unavailable");dlclose(b->library);free(b);return NULL;}
 if(bounded>0&&bounded!=3&&bounded!=4&&bounded!=7&&!b->bounded){snprintf(error,capacity,bounded==2?"reuse GPU ABI unavailable":bounded==5?"compact GPU ABI unavailable":bounded==6?"streamed GPU ABI unavailable":"bounded GPU ABI unavailable");dlclose(b->library);free(b);return NULL;}
 if(bounded==5&&!b->compact_stats){snprintf(error,capacity,"compact stats ABI unavailable");dlclose(b->library);free(b);return NULL;}
 if(bounded==6&&(!b->encode_streamed||!b->stream_stats)){snprintf(error,capacity,"streamed GPU ABI unavailable");dlclose(b->library);free(b);return NULL;}
 if(bounded==7&&!b->apply_backed){snprintf(error,capacity,"host-backed ABI unavailable");dlclose(b->library);free(b);return NULL;}
 int status=(bounded==3||bounded==4||bounded==7)?b->transform(bounded==7?(unsigned)tile:bounded==4,s,poly,delta,logs,lc,powers,pc,max_points,device_cap,&b->engine,error,capacity):bounded?b->bounded(s,poly,delta,logs,lc,powers,pc,max_points,device_cap,&b->engine,error,capacity):b->create(s,poly,delta,logs,lc,powers,pc,&b->engine,error,capacity);
 if(status){dlclose(b->library);free(b);return NULL;}return b;
}
static int bridge_check(bridge* b,const uint32_t* words,size_t count,char* error,size_t capacity){if(!b->check){snprintf(error,capacity,"plan check ABI unavailable");return 1;}return b->check(b->engine,words,count,error,capacity);}
static int bridge_add(bridge* b,const uint32_t* words,size_t count,char* error,size_t capacity){return b->add(b->engine,words,count,error,capacity);}
static int bridge_apply(bridge* b,uint32_t* words,size_t count,int inverse,double* stats,char* error,size_t capacity){return b->apply(b->engine,words,count,inverse,stats,error,capacity);}
static int bridge_product(bridge* b,const uint32_t* a,size_t ac,const uint32_t* c,size_t cc,uint32_t* out,size_t oc,size_t n,double* stats,char* error,size_t capacity){if(!b->product){snprintf(error,capacity,"resident product ABI unavailable");return 1;}return b->product(b->engine,a,ac,c,cc,out,oc,n,stats,error,capacity);}
static int bridge_set_online(bridge* b,const uint32_t* p,size_t pc,const uint32_t* spectrum,size_t n,uint32_t omega,char* error,size_t cap){if(!b->set_online){snprintf(error,cap,"online CUDA ABI unavailable");return 1;}return b->set_online(b->engine,p,pc,spectrum,n,omega,error,cap);}
static int bridge_encode_online(bridge* b,const uint32_t* message,size_t k,uint32_t* word,size_t n,double* stats,char* error,size_t cap){if(!b->encode_online){snprintf(error,cap,"online CUDA ABI unavailable");return 1;}return b->encode_online(b->engine,message,k,word,n,stats,error,cap);}
static int bridge_encode_streamed(bridge* b,const uint32_t* p,size_t pc,const uint32_t* spectrum,size_t sc,const uint32_t* message,size_t k,uint32_t* word,size_t n,size_t tile,double* stats,char* error,size_t cap){if(!b->encode_streamed){snprintf(error,cap,"streamed GPU ABI unavailable");return 1;}return b->encode_streamed(b->engine,p,pc,spectrum,sc,message,k,word,n,tile,stats,error,cap);}
static int bridge_stream_stats(bridge* b,double* values,char* error,size_t cap){if(!b->stream_stats){snprintf(error,cap,"streamed GPU ABI unavailable");return 1;}return b->stream_stats(b->engine,values,error,cap);}
static int bridge_prepare_packed(bridge* b,const uint32_t* p,size_t pc,const uint32_t* spectrum,size_t sc,const uint32_t* message,size_t k,size_t n,size_t tile,double* stats,char* error,size_t cap){if(!b->prepare_packed||!b->read_packed||!b->clear_packed){snprintf(error,cap,"packed output ABI unavailable");return 1;}return b->prepare_packed(b->engine,p,pc,spectrum,sc,message,k,n,tile,stats,error,cap);}
static int bridge_read_packed(bridge* b,size_t offset,unsigned char* out,size_t count,double* seconds,char* error,size_t cap){return b->read_packed(b->engine,offset,out,count,seconds,error,cap);}
static void bridge_clear_packed(bridge* b){if(b->clear_packed){char error[128];b->clear_packed(b->engine,error,sizeof(error));}}
static int bridge_apply_backed(bridge* b,uint32_t* data,size_t n,int inverse,double* stats,char* error,size_t cap){if(!b->apply_backed){snprintf(error,cap,"host-backed ABI unavailable");return 1;}return b->apply_backed(b->engine,data,n,inverse,stats,error,cap);}
static int bridge_compact_stats(bridge* b,double* values,char* error,size_t capacity){if(!b->compact_stats){snprintf(error,capacity,"compact stats ABI unavailable");return 1;}return b->compact_stats(b->engine,values,error,capacity);}
static int bridge_encode_backed(bridge* b,const uint32_t* p,size_t pc,const uint32_t* spectrum,const uint32_t* input,size_t k,uint32_t* arena,uint32_t* word,size_t n,uint32_t omega,double* stats,char* error,size_t capacity){if(!b->encode_backed){snprintf(error,capacity,"host-backed online ABI unavailable");return 1;}return b->encode_backed(b->engine,p,pc,spectrum,input,k,arena,word,n,omega,stats,error,capacity);}
static uint64_t bridge_owned(bridge* b){return b->owned(b->engine);}
static void bridge_close(bridge* b){if(b){b->destroy(b->engine);dlclose(b->library);free(b);}}
*/
import "C"

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// Engine serializes all calls and closes owned CUDA state explicitly. Host
// slices are borrowed only until the synchronous call returns, never retained.
type Engine struct {
	mu                            sync.Mutex
	handle                        *C.bridge
	maxPoints                     int
	fieldDegree, onlineK, onlineN int
}

// New loads the explicitly supplied optional shared library and copies
// trusted base-field tables. The caller must use a checksum-verified binary.
func New(path string, s int, polynomial, delta uint32, logs, powers []uint32) (*Engine, error) {
	return newEngine(path, s, polynomial, delta, logs, powers, 1<<22, 512<<20, 0)
}

// NewBounded explicitly admits a maximum domain and total device allocation cap.
// It requires the bounded ABI. Values remain uint32; counts can reach 2^32 on
// 64-bit hosts. The cap excludes CUDA driver/context allocations.
func NewBounded(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64) (*Engine, error) {
	return newEngine(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes, 1)
}

// NewBoundedReuse selects the experimental online-only two-work-array ABI.
// SetOnline requires two padded prefix operands to fit maxPoints. Apply/Product
// are unavailable in this mode; ordinary constructors retain their prior layout.
func NewBoundedReuse(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64) (*Engine, error) {
	return newEngine(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes, 2)
}

// NewTransformBounded is an experimental transform-only engine. directRows=false
// owns data plus full scratch; true removes scratch by using the original row
// layout and may run slower. It is an explicit choice, never an online default.
// Product and SetOnline are unavailable; caps exclude driver/context memory.
func NewTransformBounded(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64, directRows bool) (*Engine, error) {
	mode := 3
	if directRows {
		mode = 4
	}
	return newEngine(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes, mode)
}

// NewBoundedCompact is an explicit experimental online mode using one full
// uint32 arena, direct rows, bounded uint16 compaction/scatter and host expansion.
// It may run slower; s<=16 and2L<=N. Ordinary constructors are unchanged.
func NewBoundedCompact(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64) (*Engine, error) {
	return newEngine(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes, 5)
}

// NewBoundedStreamed selects explicit fast-row, two-array code streaming.
// SetOnline validates metadata without retaining host pointers/device code arrays.
// Use EncodeOnlineStreamed with trusted immutable code on every call.
func NewBoundedStreamed(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64) (*Engine, error) {
	return newEngine(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes, 6)
}

func newEngine(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64, mode int, tileOverride ...int) (*Engine, error) {
	if maxPoints < 4 || uint64(maxPoints) > uint64(1)<<32 || maxPoints&(maxPoints-1) != 0 || uint64(C.size_t(deviceBytes)) != deviceBytes {
		return nil, fmt.Errorf("GPU engine limits")
	}
	if s < 2 || s > 16 || len(logs) != 1<<uint(s) || len(powers) != 2*((1<<uint(s))-1) {
		return nil, fmt.Errorf("GPU table shape")
	}
	tile := 1 << 20
	if len(tileOverride) > 0 {
		tile = tileOverride[0]
	}
	if mode == 7 && (tile < 4 || tile > 1<<26) {
		return nil, fmt.Errorf("host-backed tile bound")
	}
	name := C.CString(path)
	defer C.free(unsafe.Pointer(name))
	var message [512]C.char
	flag := C.int(mode)
	h := C.bridge_open(name, C.uint(s), C.uint(polynomial), C.uint(delta), (*C.uint32_t)(unsafe.Pointer(&logs[0])), C.size_t(len(logs)), (*C.uint32_t)(unsafe.Pointer(&powers[0])), C.size_t(len(powers)), C.size_t(maxPoints), C.size_t(deviceBytes), C.size_t(tile), flag, &message[0], 512)
	runtime.KeepAlive(logs)
	runtime.KeepAlive(powers)
	if h == nil {
		return nil, fmt.Errorf("GPU initialization: %s", C.GoString(&message[0]))
	}
	return &Engine{handle: h, maxPoints: maxPoints, fieldDegree: s}, nil
}

// CheckPlan validates trusted matrix metadata without allocating device plans or
// workspace. It checks field/format limits, independently of engine admission.
func (e *Engine) CheckPlan(words []uint32) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.handle == nil || len(words) == 0 {
		return fmt.Errorf("GPU closed/empty plan")
	}
	var message [512]C.char
	code := C.bridge_check(e.handle, (*C.uint32_t)(unsafe.Pointer(&words[0])), C.size_t(len(words)), &message[0], 512)
	runtime.KeepAlive(words)
	if code != 0 {
		return fmt.Errorf("GPU plan check: %s", C.GoString(&message[0]))
	}
	return nil
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

// Product returns the requested low coefficients of a*b. n must be a prepared
// power-of-two domain covering the full product; truncation never shrinks that
// domain. Inputs remain unchanged and device padding is cleared on every call.
// Forward transforms, products and inverse stay resident until the prefix copy.
// Inputs must be trusted canonical packed elements of this engine's field.
func (e *Engine) Product(a, b []uint32, limit, n int) ([]uint32, [8]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var stats [8]C.double
	var timing [8]float64
	if e.handle == nil || len(a) == 0 || len(b) == 0 || n < 4 || n > e.maxPoints || n&(n-1) != 0 || len(a) > n || len(b) > n || len(a) > n-len(b)+1 || limit < 1 || limit > len(a)+len(b)-1 {
		return nil, timing, fmt.Errorf("GPU product shape/prefix")
	}
	out := make([]uint32, limit)
	var message [512]C.char
	code := C.bridge_product(e.handle, (*C.uint32_t)(unsafe.Pointer(&a[0])), C.size_t(len(a)), (*C.uint32_t)(unsafe.Pointer(&b[0])), C.size_t(len(b)), (*C.uint32_t)(unsafe.Pointer(&out[0])), C.size_t(len(out)), C.size_t(n), &stats[0], &message[0], 512)
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	runtime.KeepAlive(out)
	for i := range timing {
		timing[i] = float64(stats[i])
	}
	if code != 0 {
		return nil, timing, fmt.Errorf("GPU product: %s", C.GoString(&message[0]))
	}
	return out, timing, nil
}

// Close releases plans, buffers and the dynamic library after active calls finish.
func (e *Engine) Close() { e.mu.Lock(); defer e.mu.Unlock(); C.bridge_close(e.handle); e.handle = nil }

// SetOnline copies trusted verified P/spectrum constants to the device once.
// Streamed engines validate code and install shape only; pass code per encode.
// The primitive generator defines canonical output positions. Host slices are
// not retained, and installation fails if explicit code/workspace exceeds cap.
func (e *Engine) SetOnline(p, spectrum []uint32, omega uint32) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.handle == nil || len(p) < 2 || len(spectrum) < 4 || len(spectrum) > e.maxPoints || len(p)-1 > len(spectrum)/2 {
		return fmt.Errorf("CUDA online shape/state")
	}
	var message [512]C.char
	code := C.bridge_set_online(e.handle, (*C.uint32_t)(unsafe.Pointer(&p[0])), C.size_t(len(p)), (*C.uint32_t)(unsafe.Pointer(&spectrum[0])), C.size_t(len(spectrum)), C.uint32_t(omega), &message[0], 512)
	runtime.KeepAlive(p)
	runtime.KeepAlive(spectrum)
	if code != 0 {
		return fmt.Errorf("CUDA online setup: %s", C.GoString(&message[0]))
	}
	e.onlineK = len(p) - 1
	e.onlineN = len(spectrum)
	return nil
}

// EncodeOnline borrows message/output only until this synchronous call returns.
// All mutable work stays in engine-owned device arrays. Inputs must contain base
// field elements; native validation rejects out-of-range input before kernels.
// Output must have full domain length. Array indices match gpu_api.h stats[10].
func (e *Engine) EncodeOnline(input, word []uint32) ([10]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var stats [10]C.double
	var out [10]float64
	if e.handle == nil || len(input) == 0 || len(word) == 0 {
		return out, fmt.Errorf("CUDA online closed/empty input")
	}
	var message [512]C.char
	code := C.bridge_encode_online(e.handle, (*C.uint32_t)(unsafe.Pointer(&input[0])), C.size_t(len(input)), (*C.uint32_t)(unsafe.Pointer(&word[0])), C.size_t(len(word)), &stats[0], &message[0], 512)
	runtime.KeepAlive(input)
	runtime.KeepAlive(word)
	for i := range out {
		out[i] = float64(stats[i])
	}
	if code != 0 {
		return out, fmt.Errorf("CUDA online encode: %s", C.GoString(&message[0]))
	}
	return out, nil
}

// CompactStats returns the last completed compact call's GPU compaction seconds,
// host expansion seconds and bulk-download bytes. Calls serialize, but a caller
// must exclude other encodes to associate this snapshot with its own encode.
func (e *Engine) CompactStats() ([3]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out [3]float64
	if e.handle == nil {
		return out, fmt.Errorf("closed compact stats")
	}
	var values [3]C.double
	var message [512]C.char
	if C.bridge_compact_stats(e.handle, &values[0], &message[0], 512) != 0 {
		return out, fmt.Errorf("compact stats: %s", C.GoString(&message[0]))
	}
	for i := range out {
		out[i] = float64(values[i])
	}
	return out, nil
}

// EncodeOnlineStreamed borrows verified immutable P/spectrum only until return.
// TileElements bounds a scratch-borrowed spectrum chunk, at most 2^26 elements.
// Output must be disjoint from code/input. Code-copy time is part of H2D stats.
func (e *Engine) EncodeOnlineStreamed(input, word, p, spectrum []uint32, tileElements int) ([10]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var stats [10]C.double
	var out [10]float64
	if e.handle == nil || len(input) == 0 || len(word) == 0 || len(p) != len(input)+1 || len(spectrum) != len(word) || tileElements < 1 || tileElements > min(len(word), 1<<26) {
		return out, fmt.Errorf("streamed code shape/state/tile")
	}
	// Reject writable overlaps before C: uploads may occur after output allocation.
	start := uintptr(unsafe.Pointer(&word[0]))
	end := start + uintptr(len(word))*4
	for _, values := range [][]uint32{input, p, spectrum} {
		a := uintptr(unsafe.Pointer(&values[0]))
		if start < a+uintptr(len(values))*4 && a < end {
			return out, fmt.Errorf("streamed output overlaps input/code")
		}
	}
	var message [512]C.char
	code := C.bridge_encode_streamed(e.handle, (*C.uint32_t)(unsafe.Pointer(&p[0])), C.size_t(len(p)), (*C.uint32_t)(unsafe.Pointer(&spectrum[0])), C.size_t(len(spectrum)), (*C.uint32_t)(unsafe.Pointer(&input[0])), C.size_t(len(input)), (*C.uint32_t)(unsafe.Pointer(&word[0])), C.size_t(len(word)), C.size_t(tileElements), &stats[0], &message[0], 512)
	runtime.KeepAlive(input)
	runtime.KeepAlive(word)
	runtime.KeepAlive(p)
	runtime.KeepAlive(spectrum)
	for i := range out {
		out[i] = float64(stats[i])
	}
	if code != 0 {
		return out, fmt.Errorf("CUDA streamed encode: %s", C.GoString(&message[0]))
	}
	return out, nil
}

// StreamStats snapshots the last successful call's code-copy seconds, bytes,
// spectrum chunk count and scratch tile bytes. Exclude other encodes when reading.
func (e *Engine) StreamStats() ([4]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out [4]float64
	var values [4]C.double
	var message [512]C.char
	if e.handle == nil {
		return out, fmt.Errorf("closed stream stats")
	}
	if C.bridge_stream_stats(e.handle, &values[0], &message[0], 512) != 0 {
		return out, fmt.Errorf("CUDA stream stats: %s", C.GoString(&message[0]))
	}
	for i := range out {
		out[i] = float64(values[i])
	}
	return out, nil
}

// PackedStats describes bounded packed output. Download time is also in native
// transfer stats; WriteNS/HashNS are separate CPU stages. SHA256 is valid only on
// success; WrittenBytes can describe a partial sink after failure.
type PackedStats struct {
	WrittenBytes                uint64
	SHA256                      string
	BufferBytes                 int
	DownloadNS, WriteNS, HashNS int64
}

// EncodePackedTo writes resident-code output directly in canonical s-bit format.
// It holds the engine lock through writes; the writer must not reenter this engine.
// Failures may leave a partial sink; callers control atomic file publication.
func (e *Engine) EncodePackedTo(input []uint32, destination io.Writer, bufferBytes int) ([10]float64, PackedStats, error) {
	return e.encodePackedTo(input, nil, nil, destination, bufferBytes, 0)
}

// EncodePackedStreamedTo is the explicit streamed-code counterpart, borrowing
// immutable code for the call and using the same scratch-borrowed code tiles.
func (e *Engine) EncodePackedStreamedTo(input, p, spectrum []uint32, destination io.Writer, bufferBytes, tile int) ([10]float64, PackedStats, error) {
	return e.encodePackedTo(input, p, spectrum, destination, bufferBytes, tile)
}
func (e *Engine) encodePackedTo(input, p, spectrum []uint32, destination io.Writer, bufferBytes, tile int) ([10]float64, PackedStats, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out [10]float64
	var result PackedStats
	if e.handle == nil || e.onlineN == 0 || len(input) != e.onlineK || destination == nil || bufferBytes < 1 || bufferBytes > 64<<20 {
		return out, result, fmt.Errorf("packed output shape/state/buffer")
	}
	if p != nil && (len(p) != e.onlineK+1 || len(spectrum) != e.onlineN || tile < 1 || tile > min(e.onlineN, 1<<26)) {
		return out, result, fmt.Errorf("packed streamed code shape/tile")
	}
	ptr := func(v []uint32) *C.uint32_t {
		if len(v) == 0 {
			return nil
		}
		return (*C.uint32_t)(unsafe.Pointer(&v[0]))
	}
	var values [10]C.double
	var message [512]C.char
	code := C.bridge_prepare_packed(e.handle, ptr(p), C.size_t(len(p)), ptr(spectrum), C.size_t(len(spectrum)), ptr(input), C.size_t(len(input)), C.size_t(e.onlineN), C.size_t(tile), &values[0], &message[0], 512)
	runtime.KeepAlive(input)
	runtime.KeepAlive(p)
	runtime.KeepAlive(spectrum)
	for i := range out {
		out[i] = float64(values[i])
	}
	defer C.bridge_clear_packed(e.handle)
	if code != 0 {
		return out, result, fmt.Errorf("CUDA packed prepare: %s", C.GoString(&message[0]))
	}
	total := (uint64(e.onlineN)*uint64(e.fieldDegree) + 7) / 8
	buffer := make([]byte, min(uint64(bufferBytes), total))
	result.BufferBytes = len(buffer)
	hash := sha256.New()
	for offset := uint64(0); offset < total; {
		count := min(uint64(len(buffer)), total-offset)
		var seconds C.double
		if C.bridge_read_packed(e.handle, C.size_t(offset), (*C.uchar)(unsafe.Pointer(&buffer[0])), C.size_t(count), &seconds, &message[0], 512) != 0 {
			return out, result, fmt.Errorf("CUDA packed read: %s", C.GoString(&message[0]))
		}
		runtime.KeepAlive(buffer)
		out[5] += float64(seconds)
		out[7] += float64(seconds)
		result.DownloadNS += int64(float64(seconds) * 1e9)
		start := time.Now()
		written, err := destination.Write(buffer[:int(count)])
		result.WriteNS += time.Since(start).Nanoseconds()
		if written < 0 || written > int(count) {
			return out, result, io.ErrShortWrite
		}
		start = time.Now()
		hash.Write(buffer[:written])
		result.HashNS += time.Since(start).Nanoseconds()
		result.WrittenBytes += uint64(written)
		if err != nil {
			return out, result, err
		}
		if written != int(count) {
			return out, result, io.ErrShortWrite
		}
		offset += count
	}
	result.SHA256 = hex.EncodeToString(hash.Sum(nil))
	return out, result, nil
}

// NewHostBacked creates an explicit transform-only engine whose two device
// arrays are bounded by tileElements, not maxPoints. The full mutable domain
// stays in caller host RAM. Extra transfers and host routing can be much slower.
// AddPlan rejects tiles shorter than a full row/column. Ordinary APIs unchanged.
func NewHostBacked(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64, tileElements int) (*Engine, error) {
	return newEngine(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes, 7, tileElements)
}

// ApplyBacked transforms borrowed host data in place with GPU arithmetic and
// bounded device tiles/host column staging. No pointer is retained. Timing[12]
// follows gpu_api.h; transfers are repeated, unlike resident Apply. A CUDA
// failure may leave data partially transformed; no host pointer is retained.
func (e *Engine) ApplyBacked(data []uint32, inverse bool) ([12]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out [12]float64
	if e.handle == nil || len(data) < 4 || len(data) > e.maxPoints {
		return out, fmt.Errorf("host-backed state/shape")
	}
	flag := 0
	if inverse {
		flag = 1
	}
	var values [12]C.double
	var message [512]C.char
	code := C.bridge_apply_backed(e.handle, (*C.uint32_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), C.int(flag), &values[0], &message[0], 512)
	runtime.KeepAlive(data)
	for i := range out {
		out[i] = float64(values[i])
	}
	if code != 0 {
		return out, fmt.Errorf("GPU host-backed: %s", C.GoString(&message[0]))
	}
	return out, nil
}

// EncodeOnlineBacked runs the explicit five-transform host-backed schedule.
// All slices are borrowed synchronously. Arena/word must be separate N arrays,
// disjoint from immutable inputs. Errors may leave arena/word partially modified.
// stats[16] follows gpu_api.h. It retains neither inputs nor a full orbit map.
func (e *Engine) EncodeOnlineBacked(p, spectrum, message, arena, word []uint32, omega uint32) ([16]float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var result [16]float64
	if e.handle == nil || len(message) < 1 || len(p) != len(message)+1 || len(spectrum) != len(word) || len(arena) != len(word) || len(word) < 4 || len(word) > e.maxPoints {
		return result, fmt.Errorf("host-backed online state/shape")
	}
	// Refuse all overlap, including shifted views, before native writes.
	slices := [][]uint32{p, spectrum, message, arena, word}
	for i := 3; i < len(slices); i++ {
		for j := 0; j < i; j++ {
			a, b := slices[i], slices[j]
			first, last := uintptr(unsafe.Pointer(&a[0])), uintptr(unsafe.Pointer(&a[len(a)-1]))+4
			other, end := uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&b[len(b)-1]))+4
			if first < end && other < last {
				return result, fmt.Errorf("host-backed online buffers overlap")
			}
		}
	}
	var stats [16]C.double
	var detail [512]C.char
	ptr := func(a []uint32) *C.uint32_t { return (*C.uint32_t)(unsafe.Pointer(&a[0])) }
	code := C.bridge_encode_backed(e.handle, ptr(p), C.size_t(len(p)), ptr(spectrum), ptr(message), C.size_t(len(message)), ptr(arena), ptr(word), C.size_t(len(word)), C.uint32_t(omega), &stats[0], &detail[0], 512)
	runtime.KeepAlive(slices)
	for i := range result {
		result[i] = float64(stats[i])
	}
	if code != 0 {
		return result, fmt.Errorf("GPU host-backed online: %s", C.GoString(&detail[0]))
	}
	return result, nil
}
