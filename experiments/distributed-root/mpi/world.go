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
static int open_world(int *rank,int *size,int *provided) {
    int code=MPI_Init_thread(NULL,NULL,MPI_THREAD_FUNNELED,provided);
    if(code!=MPI_SUCCESS) return code;
    MPI_Comm_size(MPI_COMM_WORLD,size);
    if(*size<2 || *size>64 || *provided<MPI_THREAD_FUNNELED) return MPI_ERR_OTHER;
    return MPI_Comm_rank(MPI_COMM_WORLD,rank);
}
static int peer_count(int rank,uint64_t send,uint64_t *receive) {
    return MPI_Sendrecv(&send,1,MPI_UINT64_T,1-rank,0,receive,1,MPI_UINT64_T,
                        1-rank,0,MPI_COMM_WORLD,MPI_STATUS_IGNORE);
}
static int peer_words(int rank,uint32_t *send,int count,uint32_t *receive,int capacity) {
    MPI_Status status;
    int code=MPI_Sendrecv(send,count,MPI_UINT32_T,1-rank,1,receive,capacity,
                         MPI_UINT32_T,1-rank,1,MPI_COMM_WORLD,&status);
    if(code!=MPI_SUCCESS) return code;
    int actual;
    code=MPI_Get_count(&status,MPI_UINT32_T,&actual);
    return code==MPI_SUCCESS && actual!=capacity ? MPI_ERR_COUNT : code;
}
// Count/data pairs use fixed tags and exact source; the tree consumes them
// sequentially. Support routing finishes before polynomial reduction starts.
// Totals stay64bit; only bounded chunk lengths use MPI's int-count ABI.
// Payloads are direct caller-array views, without an extra staging allocation.
static void chunk_progress(int peer,int wide,int sending,uint64_t done,uint64_t total,double began,double now) {
    int rank;MPI_Comm_rank(MPI_COMM_WORLD,&rank);
    char utc[32];time_t stamp=time(NULL);struct tm tm;gmtime_r(&stamp,&tm);
    strftime(utc,sizeof(utc),"%Y-%m-%dT%H:%M:%SZ",&tm);
    fprintf(stderr,"{\"kind\":\"mpi-chunk\",\"utc\":\"%s\",\"rank\":%d,\"peer\":%d,\"phase\":\"%s\",\"sending\":%s,\"event\":\"%s\",\"completed\":%llu,\"total\":%llu,\"unit\":\"words\",\"elapsed_seconds\":%.6f}\n",
        utc,rank,peer,wide?"support-transfer":"root-transfer",sending?"true":"false",done==total?"completed":"progress",
        (unsigned long long)done,(unsigned long long)total,now-began);
    fflush(stderr);
}
static int send_values(int peer,void *values,uint64_t count,int wide,int chunk) {
    int tag=wide?10:20;
    int code=MPI_Send(&count,1,MPI_UINT64_T,peer,tag,MPI_COMM_WORLD);
    if(code!=MPI_SUCCESS || count==0)return code;
    double began=MPI_Wtime(),last=began;
    for(uint64_t first=0;first<count;) {
        int take=(int)((count-first)<(uint64_t)chunk?count-first:(uint64_t)chunk);
        code=MPI_Send((char*)values+first*(wide?8:4),take,wide?MPI_UINT64_T:MPI_UINT32_T,peer,tag+1,MPI_COMM_WORLD);
        if(code!=MPI_SUCCESS)return code;
        first+=(uint64_t)take;double now=MPI_Wtime();
        if(count>(uint64_t)chunk && (now-last>=30 || first==count)) {chunk_progress(peer,wide,1,first,count,began,now);last=now;}
    }
    return MPI_SUCCESS;
}
static int receive_count(int peer,uint64_t *count,int wide) {
    return MPI_Recv(count,1,MPI_UINT64_T,peer,wide?10:20,MPI_COMM_WORLD,MPI_STATUS_IGNORE);
}
static int receive_values(int peer,void *values,uint64_t count,int wide,int chunk) {
    double began=MPI_Wtime(),last=began;
    for(uint64_t first=0;first<count;) {
        int take=(int)((count-first)<(uint64_t)chunk?count-first:(uint64_t)chunk);MPI_Status status;
        int code=MPI_Recv((char*)values+first*(wide?8:4),take,wide?MPI_UINT64_T:MPI_UINT32_T,peer,wide?11:21,MPI_COMM_WORLD,&status);
        if(code!=MPI_SUCCESS)return code;
        int actual;code=MPI_Get_count(&status,wide?MPI_UINT64_T:MPI_UINT32_T,&actual);
        if(code!=MPI_SUCCESS || actual!=take)return MPI_ERR_COUNT;
        first+=(uint64_t)take;double now=MPI_Wtime();
        if(count>(uint64_t)chunk && (now-last>=30 || first==count)) {chunk_progress(peer,wide,0,first,count,began,now);last=now;}
    }
    return MPI_SUCCESS;
}
// Diagnostic control exchange uses a distinct tag from all data messages.
static int exchange_proof(int peer,unsigned char *send,unsigned char *receive) {
    MPI_Status status;
    int code=MPI_Sendrecv(send,64,MPI_BYTE,peer,30,receive,64,MPI_BYTE,peer,30,MPI_COMM_WORLD,&status);
    if(code!=MPI_SUCCESS)return code;
    int actual;code=MPI_Get_count(&status,MPI_BYTE,&actual);
    return code==MPI_SUCCESS && actual!=64?MPI_ERR_COUNT:code;
}
static uint64_t share_count(uint64_t count) {
    if(MPI_Bcast(&count,1,MPI_UINT64_T,0,MPI_COMM_WORLD)!=MPI_SUCCESS)MPI_Abort(MPI_COMM_WORLD,93);
    return count;
}
static double max_seconds(double value) {
    double result=0;
    if(MPI_Allreduce(&value,&result,1,MPI_DOUBLE,MPI_MAX,MPI_COMM_WORLD)!=MPI_SUCCESS)
        MPI_Abort(MPI_COMM_WORLD,91);
    return result;
}
static void abort_world(void) {MPI_Abort(MPI_COMM_WORLD,92);}
static int barrier_world(void) {return MPI_Barrier(MPI_COMM_WORLD);}
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
	Rank, Size int
	// ChunkWords bounds each MPI int-count operation; metadata totals stay64bit.
	// All ranks must choose the same1..2^20 value. Default2^20 words; bounded
	// tests explicitly use smaller chunks to cover boundaries without huge RAM.
	ChunkWords int
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
	return &World{Rank: int(rank), Size: int(size), ChunkWords: 1 << 20}, nil
}

// Abort terminates both ranks after a local computation/validation failure.
func (w *World) Abort() { C.abort_world() }

// Close finalizes MPI and releases the calling thread. It follows all work.
func (w *World) Close() error {
	code := C.MPI_Finalize()
	runtime.UnlockOSThread()
	if code != C.MPI_SUCCESS {
		return fmt.Errorf("MPI finalize: %d", int(code))
	}
	return nil
}

// Barrier excludes earlier rank skew from a subsequent local pipeline clock.
func (w *World) Barrier() {
	if C.barrier_world() != C.MPI_SUCCESS {
		w.Abort()
	}
}

// Max reports a slowest-rank duration; call it after stopping the local clock.
func (w *World) Max(seconds float64) float64 { return float64(C.max_seconds(C.double(seconds))) }

// Peer exchanges complete uint32 coefficient arrays with the other rank.
// Both local and remote lengths are bounded before allocation or MPI int casts.
// This synchronous call borrows slices only until Sendrecv returns.
func (w *World) Peer(values []uint32, maxWords int) ([]uint32, error) {
	fail := func() ([]uint32, error) { w.Abort(); return nil, fmt.Errorf("MPI coefficient exchange") }
	if w.Size != 2 || len(values) == 0 || len(values) > maxWords || maxWords > 1<<31-1 {
		return fail()
	}
	var count C.uint64_t
	if C.peer_count(C.int(w.Rank), C.uint64_t(len(values)), &count) != C.MPI_SUCCESS {
		return fail()
	}
	if count == 0 || uint64(count) > uint64(maxWords) {
		return fail()
	}
	received := make([]uint32, int(count))
	if C.peer_words(C.int(w.Rank), (*C.uint32_t)(unsafe.Pointer(&values[0])), C.int(len(values)),
		(*C.uint32_t)(unsafe.Pointer(&received[0])), C.int(len(received))) != C.MPI_SUCCESS {
		return fail()
	}
	runtime.KeepAlive(values)
	return received, nil
}

// ShareCount broadcasts bounded case metadata from coordinator rank0. It does
// not broadcast support or polynomial arrays.
func (w *World) ShareCount(count uint64) uint64 { return uint64(C.share_count(C.uint64_t(count))) }

// SendIndices routes a borrowed uint64 support packet; an empty packet closes
// that owner's stream. Packet size is capped independently of total support.
func (w *World) SendIndices(peer int, values []uint64) error {
	if peer < 0 || peer >= w.Size || peer == w.Rank || len(values) > 16384 || w.ChunkWords < 1 || w.ChunkWords > 1<<20 {
		w.Abort()
		return fmt.Errorf("support send bounds")
	}
	var ptr unsafe.Pointer
	if len(values) > 0 {
		ptr = unsafe.Pointer(&values[0])
	}
	if C.send_values(C.int(peer), ptr, C.uint64_t(len(values)), 1, C.int(w.ChunkWords)) != C.MPI_SUCCESS {
		w.Abort()
		return fmt.Errorf("support send")
	}
	runtime.KeepAlive(values)
	return nil
}

// ReceiveIndices receives one packet from the coordinator. Empty means end.
func (w *World) ReceiveIndices() ([]uint64, error) {
	var count C.uint64_t
	if w.ChunkWords < 1 || w.ChunkWords > 1<<20 || w.Rank == 0 || C.receive_count(0, &count, 1) != C.MPI_SUCCESS || count > 16384 {
		w.Abort()
		return nil, fmt.Errorf("support count")
	}
	if count == 0 {
		return nil, nil
	}
	values := make([]uint64, int(count))
	if C.receive_values(0, unsafe.Pointer(&values[0]), C.uint64_t(count), 1, C.int(w.ChunkWords)) != C.MPI_SUCCESS {
		w.Abort()
		return nil, fmt.Errorf("support payload")
	}
	return values, nil
}

// SendRoot hands a complete normalized polynomial to its next tree owner.
func (w *World) SendRoot(peer int, values []uint32, maxWords int) error {
	if peer < 0 || peer >= w.Size || peer == w.Rank || len(values) == 0 || len(values) > maxWords || w.ChunkWords < 1 || w.ChunkWords > 1<<20 {
		w.Abort()
		return fmt.Errorf("root send bounds")
	}
	if C.send_values(C.int(peer), unsafe.Pointer(&values[0]), C.uint64_t(len(values)), 0, C.int(w.ChunkWords)) != C.MPI_SUCCESS {
		w.Abort()
		return fmt.Errorf("root send")
	}
	runtime.KeepAlive(values)
	return nil
}

// ReceiveRoot checks the incoming length before allocating or MPI int casts.
func (w *World) ReceiveRoot(peer, maxWords int) ([]uint32, error) {
	var count C.uint64_t
	if peer < 0 || peer >= w.Size || peer == w.Rank || maxWords < 1 || w.ChunkWords < 1 || w.ChunkWords > 1<<20 {
		w.Abort()
		return nil, fmt.Errorf("root receive bounds")
	}
	if C.receive_count(C.int(peer), &count, 0) != C.MPI_SUCCESS || count == 0 || uint64(count) > uint64(maxWords) {
		w.Abort()
		return nil, fmt.Errorf("root receive count")
	}
	values := make([]uint32, int(count))
	if C.receive_values(C.int(peer), unsafe.Pointer(&values[0]), C.uint64_t(count), 0, C.int(w.ChunkWords)) != C.MPI_SUCCESS {
		w.Abort()
		return nil, fmt.Errorf("root receive payload")
	}
	return values, nil
}

// ExchangeProof is test-only control telemetry, not an adversarial authentication
// protocol. It exchanges fixed pre/post fingerprints without large data buffers.
func (w *World) ExchangeProof(peer int, proof [64]byte) ([64]byte, error) {
	var remote [64]byte
	if peer < 0 || peer >= w.Size || peer == w.Rank {
		w.Abort()
		return remote, fmt.Errorf("proof peer bounds")
	}
	if C.exchange_proof(C.int(peer), (*C.uchar)(unsafe.Pointer(&proof[0])), (*C.uchar)(unsafe.Pointer(&remote[0]))) != C.MPI_SUCCESS {
		w.Abort()
		return remote, fmt.Errorf("proof exchange")
	}
	return remote, nil
}
