//go:build liftrs_mpi && cgo

// Package mpi is the bounded rank transport used by the distributed-root
// experiment. It is not a distributed preprocessing API. MPI is required only
// when explicitly building with liftrs_mpi and cgo using CC=mpicc/CXX=mpicxx.
package mpi

/*
#include <mpi.h>
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <time.h>
#include <string.h>
// Explicit communicator handles isolate concurrent subtree traffic even when
// operation tags/identifiers repeat in disjoint groups. No global active group.
typedef struct { MPI_Comm value; } depir_comm;
static depir_comm world_handle;
static depir_comm *world_comm(void) {world_handle.value=MPI_COMM_WORLD;return &world_handle;}
static int split_comm(depir_comm *parent,int color,int key,depir_comm **result,int *rank,int *size) {
    depir_comm *child=(depir_comm*)calloc(1,sizeof(depir_comm));
    if(!child)return MPI_ERR_NO_MEM;
    int code=MPI_Comm_split(parent->value,color,key,&child->value);
    if(code!=MPI_SUCCESS){free(child);return code;}
    code=MPI_Comm_rank(child->value,rank);
    if(code==MPI_SUCCESS)code=MPI_Comm_size(child->value,size);
    if(code!=MPI_SUCCESS){MPI_Comm_free(&child->value);free(child);return code;}
    *result=child;return MPI_SUCCESS;
}
static int close_comm(depir_comm *child) {
    int code=MPI_Comm_free(&child->value);if(code==MPI_SUCCESS)free(child);return code;
}
static int all_counts(depir_comm *context,uint64_t value,uint64_t *result) {
    return MPI_Allgather(&value,1,MPI_UINT64_T,result,1,MPI_UINT64_T,context->value);
}
static int open_world(int *rank,int *size,int *provided) {
    int code=MPI_Init_thread(NULL,NULL,MPI_THREAD_FUNNELED,provided);
    if(code!=MPI_SUCCESS) return code;
    MPI_Comm_size(MPI_COMM_WORLD,size);
    if(*size<2 || *size>64 || *provided<MPI_THREAD_FUNNELED) return MPI_ERR_OTHER;
    return MPI_Comm_rank(MPI_COMM_WORLD,rank);
}
static int peer_count(depir_comm *context,int rank,uint64_t send,uint64_t *receive) {
    return MPI_Sendrecv(&send,1,MPI_UINT64_T,1-rank,0,receive,1,MPI_UINT64_T,
                        1-rank,0,context->value,MPI_STATUS_IGNORE);
}
static int peer_words(depir_comm *context,int rank,uint32_t *send,int count,uint32_t *receive,int capacity) {
    MPI_Status status;
    int code=MPI_Sendrecv(send,count,MPI_UINT32_T,1-rank,1,receive,capacity,
                         MPI_UINT32_T,1-rank,1,context->value,&status);
    if(code!=MPI_SUCCESS) return code;
    int actual;
    code=MPI_Get_count(&status,MPI_UINT32_T,&actual);
    return code==MPI_SUCCESS && actual!=capacity ? MPI_ERR_COUNT : code;
}
// Count/data pairs use fixed tags and exact source; the tree consumes them
// sequentially. Support routing finishes before polynomial reduction starts.
// Totals stay64bit; only bounded chunk lengths use MPI's int-count ABI.
// Optional MPI-owned staging copies one bounded chunk; direct views remain the default.
static void chunk_progress(depir_comm *context,int peer,int tag,int sending,uint64_t done,uint64_t total,double began,double now) {
    int rank,group_rank,group_size;MPI_Comm_rank(MPI_COMM_WORLD,&rank);
    MPI_Comm_rank(context->value,&group_rank);MPI_Comm_size(context->value,&group_size);
    char utc[32];time_t stamp=time(NULL);struct tm tm;gmtime_r(&stamp,&tm);
    strftime(utc,sizeof(utc),"%Y-%m-%dT%H:%M:%SZ",&tm);
    fprintf(stderr,"{\"kind\":\"mpi-chunk\",\"utc\":\"%s\",\"rank\":%d,\"group_rank\":%d,\"group_size\":%d,\"peer\":%d,\"phase\":\"%s\",\"sending\":%s,\"event\":\"%s\",\"completed\":%llu,\"total\":%llu,\"unit\":\"words\",\"elapsed_seconds\":%.6f}\n",
        utc,rank,group_rank,group_size,peer,tag==10?"support-transfer":tag==20?"root-transfer":tag==40?"coefficient32-transfer":"coefficient64-transfer",sending?"true":"false",done==total?"completed":"progress",
        (unsigned long long)done,(unsigned long long)total,now-began);
    fflush(stderr);
}
static int send_values(depir_comm *context,int peer,void *values,uint64_t count,int wide,int chunk,void *staging,int tag) {
    int code=MPI_Send(&count,1,MPI_UINT64_T,peer,tag,context->value);
    if(code!=MPI_SUCCESS || count==0)return code;
    double began=MPI_Wtime(),last=began;
    for(uint64_t first=0;first<count;) {
        int take=(int)((count-first)<(uint64_t)chunk?count-first:(uint64_t)chunk);
        void *source=(char*)values+first*(wide?8:4);
        if(staging){memcpy(staging,source,(size_t)take*(wide?8:4));source=staging;}
        code=MPI_Send(source,take,wide?MPI_UINT64_T:MPI_UINT32_T,peer,tag+1,context->value);
        if(code!=MPI_SUCCESS)return code;
        first+=(uint64_t)take;double now=MPI_Wtime();
        if(count>(uint64_t)chunk && (now-last>=30 || first==count)) {chunk_progress(context,peer,tag,1,first,count,began,now);last=now;}
    }
    return MPI_SUCCESS;
}
static int receive_count(depir_comm *context,int peer,uint64_t *count,int tag) {
    uint64_t value=0;
    int code=MPI_Recv(&value,1,MPI_UINT64_T,peer,tag,context->value,MPI_STATUS_IGNORE);
    if(code==MPI_SUCCESS)*count=value;
    return code;
}
static int receive_values(depir_comm *context,int peer,void *values,uint64_t count,int wide,int chunk,void *staging,int tag) {
    double began=MPI_Wtime(),last=began;
    for(uint64_t first=0;first<count;) {
        int take=(int)((count-first)<(uint64_t)chunk?count-first:(uint64_t)chunk);MPI_Status status;
        void *destination=(char*)values+first*(wide?8:4);
        int code=MPI_Recv(staging?staging:destination,take,wide?MPI_UINT64_T:MPI_UINT32_T,peer,tag+1,context->value,&status);
        if(code!=MPI_SUCCESS)return code;
        int actual;code=MPI_Get_count(&status,wide?MPI_UINT64_T:MPI_UINT32_T,&actual);
        if(code!=MPI_SUCCESS || actual!=take)return MPI_ERR_COUNT;
        if(staging)memcpy(destination,staging,(size_t)take*(wide?8:4));
        first+=(uint64_t)take;double now=MPI_Wtime();
        if(count>(uint64_t)chunk && (now-last>=30 || first==count)) {chunk_progress(context,peer,tag,0,first,count,began,now);last=now;}
    }
    return MPI_SUCCESS;
}
// Diagnostic control exchange uses a distinct tag from all data messages.
static int exchange_proof(depir_comm *context,int peer,unsigned char *send,unsigned char *receive) {
    MPI_Status status;unsigned char source[64],destination[64];memcpy(source,send,64);
    int code=MPI_Sendrecv(source,64,MPI_BYTE,peer,30,destination,64,MPI_BYTE,peer,30,context->value,&status);
    if(code!=MPI_SUCCESS)return code;
    int actual;code=MPI_Get_count(&status,MPI_BYTE,&actual);
    if(code!=MPI_SUCCESS || actual!=64)return MPI_ERR_COUNT;
    memcpy(receive,destination,64);return MPI_SUCCESS;
}
static void *allocate_staging(uint64_t bytes) {
    void *buffer=NULL;
    if(MPI_Alloc_mem((MPI_Aint)bytes,MPI_INFO_NULL,&buffer)!=MPI_SUCCESS)return NULL;
    return buffer;
}
static int release_staging(void *buffer){return MPI_Free_mem(buffer);}
static uint64_t share_count(depir_comm *context,uint64_t count) {
    if(MPI_Bcast(&count,1,MPI_UINT64_T,0,context->value)!=MPI_SUCCESS)MPI_Abort(MPI_COMM_WORLD,93);
    return count;
}
static double max_seconds(depir_comm *context,double value) {
    double result=0;
    if(MPI_Allreduce(&value,&result,1,MPI_DOUBLE,MPI_MAX,context->value)!=MPI_SUCCESS)
        MPI_Abort(MPI_COMM_WORLD,91);
    return result;
}
static void abort_world(void) {MPI_Abort(MPI_COMM_WORLD,92);}
static int barrier_world(depir_comm *context) {return MPI_Barrier(context->value);}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// World keeps MPI on the initializing OS thread (MPI_THREAD_FUNNELED). The caller
// must invoke every method, including Close, from the opening goroutine. CUDA
// calls may use other threads; MPI never sees a device or retained Go pointer.
type World struct {
	// Rank/Size are read-only communicator-local metadata.
	Rank, Size int
	comm       *C.depir_comm
	parent     *World
	children   int
	// ChunkWords bounds each MPI int-count operation; metadata totals stay64bit.
	// All ranks must choose the same1..2^20 value. Default2^20 words; bounded
	// tests explicitly use smaller chunks to cover boundaries without huge RAM.
	ChunkWords   int
	staging      unsafe.Pointer
	stagingWords int
}

// Open initializes2..64 ranks. A failed distributed operation aborts the
// communicator, so a peer cannot hang waiting for a rank that returned an error.
func Open() (*World, error) {
	runtime.LockOSThread()
	var rank, size, provided C.int
	if code := C.open_world(&rank, &size, &provided); code != C.MPI_SUCCESS {
		C.abort_world()
		return nil, fmt.Errorf("MPI initialization: %d", int(code))
	}
	return &World{Rank: int(rank), Size: int(size), ChunkWords: 1 << 20, comm: C.world_comm()}, nil
}

// Abort terminates all ranks after a local computation/validation failure.
func (w *World) Abort() { C.abort_world() }

// Close frees a child communicator, or finalizes MPI/releases the root thread.
// Close children before parents; every member must close its group collectively.
func (w *World) Close() error {
	if !w.live() || w.children != 0 {
		return fmt.Errorf("MPI close requires live communicator without children")
	}
	if w.staging != nil {
		if C.release_staging(w.staging) != C.MPI_SUCCESS {
			w.Abort()
			return fmt.Errorf("MPI staging release")
		}
		w.staging = nil
	}
	if w.parent != nil {
		if code := C.close_comm(w.comm); code != C.MPI_SUCCESS {
			w.Abort()
			return fmt.Errorf("MPI communicator release: %d", int(code))
		}
		w.comm = nil
		w.parent.children--
		return nil
	}
	code := C.MPI_Finalize()
	w.comm = nil
	runtime.UnlockOSThread()
	if code != C.MPI_SUCCESS {
		return fmt.Errorf("MPI finalize: %d", int(code))
	}
	return nil
}

// Barrier excludes earlier rank skew from a subsequent local pipeline clock.
func (w *World) Barrier() {
	if !w.live() {
		panic("closed MPI communicator")
	}
	if C.barrier_world(w.comm) != C.MPI_SUCCESS {
		w.Abort()
	}
}

// Max reports a slowest-rank duration; call it after stopping the local clock.
func (w *World) Max(seconds float64) float64 {
	if !w.live() {
		panic("closed MPI communicator")
	}
	return float64(C.max_seconds(w.comm, C.double(seconds)))
}

// Peer exchanges complete uint32 coefficient arrays with the other rank.
// Both local and remote lengths are bounded before allocation or MPI int casts.
// This synchronous call borrows slices only until Sendrecv returns.
func (w *World) Peer(values []uint32, maxWords int) ([]uint32, error) {
	fail := func() ([]uint32, error) { w.Abort(); return nil, fmt.Errorf("MPI coefficient exchange") }
	if !w.live() || w.Size != 2 || len(values) == 0 || len(values) > maxWords || maxWords > 1<<31-1 {
		return fail()
	}
	var count C.uint64_t
	if C.peer_count(w.comm, C.int(w.Rank), C.uint64_t(len(values)), &count) != C.MPI_SUCCESS {
		return fail()
	}
	if count == 0 || uint64(count) > uint64(maxWords) {
		return fail()
	}
	received := make([]uint32, int(count))
	if C.peer_words(w.comm, C.int(w.Rank), (*C.uint32_t)(unsafe.Pointer(&values[0])), C.int(len(values)),
		(*C.uint32_t)(unsafe.Pointer(&received[0])), C.int(len(received))) != C.MPI_SUCCESS {
		return fail()
	}
	runtime.KeepAlive(values)
	return received, nil
}

// ShareCount broadcasts bounded case metadata from coordinator rank0. It does
// not broadcast support or polynomial arrays.
func (w *World) ShareCount(count uint64) uint64 {
	if !w.live() {
		panic("closed MPI communicator")
	}
	return uint64(C.share_count(w.comm, C.uint64_t(count)))
}

// SendIndices routes a borrowed uint64 support packet; an empty packet closes
// that owner's stream. Packet size is capped independently of total support.
func (w *World) SendIndices(peer int, values []uint64) error {
	if peer < 0 || peer >= w.Size || peer == w.Rank || len(values) > 16384 || !w.validChunks() {
		w.Abort()
		return fmt.Errorf("support send bounds")
	}
	var ptr unsafe.Pointer
	if len(values) > 0 {
		ptr = unsafe.Pointer(&values[0])
	}
	if C.send_values(w.comm, C.int(peer), ptr, C.uint64_t(len(values)), 1, C.int(w.ChunkWords), w.staging, 10) != C.MPI_SUCCESS {
		w.Abort()
		return fmt.Errorf("support send")
	}
	runtime.KeepAlive(values)
	return nil
}

// ReceiveIndices receives one packet from the coordinator. Empty means end.
func (w *World) ReceiveIndices() ([]uint64, error) {
	var count C.uint64_t
	if !w.validChunks() || w.Rank == 0 || C.receive_count(w.comm, 0, &count, 10) != C.MPI_SUCCESS || count > 16384 {
		w.Abort()
		return nil, fmt.Errorf("support count")
	}
	if count == 0 {
		return nil, nil
	}
	values := make([]uint64, int(count))
	if C.receive_values(w.comm, 0, unsafe.Pointer(&values[0]), C.uint64_t(count), 1, C.int(w.ChunkWords), w.staging, 10) != C.MPI_SUCCESS {
		w.Abort()
		return nil, fmt.Errorf("support payload")
	}
	return values, nil
}

// SendRoot hands a complete normalized polynomial to its next tree owner.
func (w *World) SendRoot(peer int, values []uint32, maxWords int) error {
	if peer < 0 || peer >= w.Size || peer == w.Rank || len(values) == 0 || len(values) > maxWords || !w.validChunks() {
		w.Abort()
		return fmt.Errorf("root send bounds")
	}
	if C.send_values(w.comm, C.int(peer), unsafe.Pointer(&values[0]), C.uint64_t(len(values)), 0, C.int(w.ChunkWords), w.staging, 20) != C.MPI_SUCCESS {
		w.Abort()
		return fmt.Errorf("root send")
	}
	runtime.KeepAlive(values)
	return nil
}

// ReceiveRoot checks the incoming length before allocating or MPI int casts.
func (w *World) ReceiveRoot(peer, maxWords int) ([]uint32, error) {
	var count C.uint64_t
	if peer < 0 || peer >= w.Size || peer == w.Rank || maxWords < 1 || !w.validChunks() {
		w.Abort()
		return nil, fmt.Errorf("root receive bounds")
	}
	if C.receive_count(w.comm, C.int(peer), &count, 20) != C.MPI_SUCCESS || count == 0 || uint64(count) > uint64(maxWords) {
		w.Abort()
		return nil, fmt.Errorf("root receive count")
	}
	values := make([]uint32, int(count))
	if C.receive_values(w.comm, C.int(peer), unsafe.Pointer(&values[0]), C.uint64_t(count), 0, C.int(w.ChunkWords), w.staging, 20) != C.MPI_SUCCESS {
		w.Abort()
		return nil, fmt.Errorf("root receive payload")
	}
	return values, nil
}

// ExchangeControl exchanges a synchronous fixed64-byte control record on tag30.
// Callers check its meaning (descriptor or pre/post proof). It is not adversarial
// authentication and does not retain Go buffers. All ranks must agree on order.
func (w *World) ExchangeControl(peer int, proof [64]byte) ([64]byte, error) {
	var remote [64]byte
	if !w.live() || peer < 0 || peer >= w.Size || peer == w.Rank {
		w.Abort()
		return remote, fmt.Errorf("proof peer bounds")
	}
	if C.exchange_proof(w.comm, C.int(peer), (*C.uchar)(unsafe.Pointer(&proof[0])), (*C.uchar)(unsafe.Pointer(&remote[0]))) != C.MPI_SUCCESS {
		w.Abort()
		return remote, fmt.Errorf("proof exchange")
	}
	return remote, nil
}

// ControlFrames gathers at most64 fixed-size control records (4KiB/rank).
// All ranks must call in the same phase order; callers bind and validate record
// contents. Lexicographic pair order prevents cycles with blocking exchanges.
func (w *World) ControlFrames(own [64]byte) ([][64]byte, error) {
	if !w.live() || w.Size < 1 || w.Size > 64 || w.Rank < 0 || w.Rank >= w.Size {
		return nil, fmt.Errorf("MPI control ownership")
	}
	out := make([][64]byte, w.Size)
	out[w.Rank] = own
	for source := 0; source < w.Size; source++ {
		for target := source + 1; target < w.Size; target++ {
			if w.Rank != source && w.Rank != target {
				continue
			}
			peer := source
			if w.Rank == source {
				peer = target
			}
			remote, err := w.ExchangeControl(peer, own)
			if err != nil {
				return nil, err
			}
			out[peer] = remote
		}
	}
	return out, nil
}

// ExchangeProof retains the original diagnostic handshake on control tag30.
func (w *World) ExchangeProof(peer int, proof [64]byte) ([64]byte, error) {
	return w.ExchangeControl(peer, proof)
}

// Coefficient tiles use distinct tag pairs40/41 and50/51. The bounded caller
// buffer is borrowed only until the blocking call returns; no receive allocation.
// Operation/shape/offset descriptors must be checked through ExchangeControl
// before calling these methods. These lanes are not concurrent MPI streams.
func (w *World) tileSend(peer int, pointer unsafe.Pointer, words int, wide bool) error {
	if peer < 0 || peer >= w.Size || peer == w.Rank || words < 1 || words > 1<<20 || !w.validChunks() {
		w.Abort()
		return fmt.Errorf("coefficient tile send bounds")
	}
	width, tag := 0, 40
	if wide {
		width, tag = 1, 50
	}
	if C.send_values(w.comm, C.int(peer), pointer, C.uint64_t(words), C.int(width), C.int(w.ChunkWords), w.staging, C.int(tag)) != C.MPI_SUCCESS {
		w.Abort()
		return fmt.Errorf("coefficient tile send")
	}
	return nil
}
func (w *World) tileReceive(peer int, pointer unsafe.Pointer, words int, wide bool) error {
	if peer < 0 || peer >= w.Size || peer == w.Rank || words < 1 || words > 1<<20 || !w.validChunks() {
		w.Abort()
		return fmt.Errorf("coefficient tile receive bounds")
	}
	width, tag := 0, 40
	if wide {
		width, tag = 1, 50
	}
	var count C.uint64_t
	if C.receive_count(w.comm, C.int(peer), &count, C.int(tag)) != C.MPI_SUCCESS || uint64(count) != uint64(words) {
		w.Abort()
		return fmt.Errorf("coefficient tile count")
	}
	if C.receive_values(w.comm, C.int(peer), pointer, C.uint64_t(count), C.int(width), C.int(w.ChunkWords), w.staging, C.int(tag)) != C.MPI_SUCCESS {
		w.Abort()
		return fmt.Errorf("coefficient tile receive")
	}
	return nil
}

// SendCoefficients32 sends a nonempty bounded base-field storage tile.
func (w *World) SendCoefficients32(peer int, values []uint32) error {
	if len(values) == 0 {
		w.Abort()
		return fmt.Errorf("empty coefficient tile")
	}
	err := w.tileSend(peer, unsafe.Pointer(&values[0]), len(values), false)
	runtime.KeepAlive(values)
	return err
}

// ReceiveCoefficients32 fills exactly the borrowed tile, without another array.
func (w *World) ReceiveCoefficients32(peer int, values []uint32) error {
	if len(values) == 0 {
		w.Abort()
		return fmt.Errorf("empty coefficient tile")
	}
	err := w.tileReceive(peer, unsafe.Pointer(&values[0]), len(values), false)
	runtime.KeepAlive(values)
	return err
}

// SendCoefficients64 preserves all bits of an extension-field storage tile.
func (w *World) SendCoefficients64(peer int, values []uint64) error {
	if len(values) == 0 {
		w.Abort()
		return fmt.Errorf("empty coefficient tile")
	}
	err := w.tileSend(peer, unsafe.Pointer(&values[0]), len(values), true)
	runtime.KeepAlive(values)
	return err
}

// ReceiveCoefficients64 fills exactly the borrowed wide tile.
func (w *World) ReceiveCoefficients64(peer int, values []uint64) error {
	if len(values) == 0 {
		w.Abort()
		return fmt.Errorf("empty coefficient tile")
	}
	err := w.tileReceive(peer, unsafe.Pointer(&values[0]), len(values), true)
	runtime.KeepAlive(values)
	return err
}

// Changing the chunk after allocation could overrun the reusable scratch buffer.
func (w *World) validChunks() bool {
	return w.live() && w.ChunkWords >= 1 && w.ChunkWords <= 1<<20 && (w.staging == nil || w.stagingWords == w.ChunkWords)
}

// EnableStaging is explicit: one MPI-owned scratch chunk, reused synchronously
// for both directions. It adds a CPU copy per payload and is not a default.
// Do not copy a live World or change ChunkWords after enabling it.
func (w *World) EnableStaging() error {
	if !w.live() || w.ChunkWords < 1 || w.ChunkWords > 1<<20 || w.staging != nil {
		return fmt.Errorf("MPI staging configuration")
	}
	w.staging = C.allocate_staging(C.uint64_t(w.ChunkWords) * 8)
	if w.staging == nil {
		w.Abort()
		return fmt.Errorf("MPI staging allocation")
	}
	w.stagingWords = w.ChunkWords
	return nil
}

// StagingBytes counts owned scratch memory, not MPI/driver internal buffers.
func (w *World) StagingBytes() uint64 {
	if w.staging == nil {
		return 0
	}
	return uint64(w.stagingWords) * 8
}

func (w *World) live() bool { return w != nil && w.comm != nil }

// Split collectively creates a group with local ranks ordered by key. Groups
// borrow the root's locked thread,own their communicator/staging and must Close
// before their parent. MPI failures abort the entire original job,not one subtree.
func (w *World) Split(color, key int) (*World, error) {
	if !w.live() || color < 0 || key < 0 || color > 1<<31-1 || key > 1<<31-1 {
		return nil, fmt.Errorf("MPI split arguments")
	}
	var rank, size C.int
	var handle *C.depir_comm
	if code := C.split_comm(w.comm, C.int(color), C.int(key), &handle, &rank, &size); code != C.MPI_SUCCESS {
		w.Abort()
		return nil, fmt.Errorf("MPI split: %d", int(code))
	}
	w.children++
	return &World{Rank: int(rank), Size: int(size), ChunkWords: w.ChunkWords, comm: handle, parent: w}, nil
}

// Counts returns one exact integer from every member,in communicator-rank order.
// Metadata is bounded to64 members;no polynomial/support data is gathered.
func (w *World) Counts(value uint64) ([]uint64, error) {
	if !w.live() || w.Size < 1 || w.Size > 64 {
		return nil, fmt.Errorf("MPI count ownership")
	}
	out := make([]uint64, w.Size)
	if code := C.all_counts(w.comm, C.uint64_t(value), (*C.uint64_t)(unsafe.Pointer(&out[0]))); code != C.MPI_SUCCESS {
		w.Abort()
		return nil, fmt.Errorf("MPI count gather: %d", int(code))
	}
	return out, nil
}
