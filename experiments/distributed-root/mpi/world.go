//go:build liftrs_mpi && cgo

// Package mpi is the bounded two-rank transport used by the distributed-root
// experiment. It is not a distributed preprocessing API. MPI is required only
// when explicitly building with liftrs_mpi and cgo using CC=mpicc/CXX=mpicxx.
package mpi

/*
#include <mpi.h>
#include <stdint.h>
#include <stdlib.h>
static int open_world(int *rank, int *provided) {
    int code=MPI_Init_thread(NULL,NULL,MPI_THREAD_FUNNELED,provided);
    if(code!=MPI_SUCCESS) return code;
    int size;
    MPI_Comm_size(MPI_COMM_WORLD,&size);
    if(size!=2 || *provided<MPI_THREAD_FUNNELED) return MPI_ERR_OTHER;
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
type World struct{ Rank int }

// Open initializes exactly two ranks. A failed distributed operation aborts the
// communicator, so a peer cannot hang waiting for a rank that returned an error.
func Open() (*World, error) {
	runtime.LockOSThread()
	var rank, provided C.int
	if code := C.open_world(&rank, &provided); code != C.MPI_SUCCESS {
		C.abort_world()
		return nil, fmt.Errorf("MPI initialization: %d", int(code))
	}
	return &World{int(rank)}, nil
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
	if len(values) == 0 || len(values) > maxWords || maxWords > 1<<31-1 {
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
