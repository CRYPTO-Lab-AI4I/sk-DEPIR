#!/usr/bin/env python3
"""Source-based allocation/traffic projections;never admits or allocates HPC work."""
from pathlib import Path
import argparse
import hashlib
import json
import subprocess
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[1]
KNOWN_K={16:918672421,17:3876566203,18:16300262095}
SOURCES=('encoding/liftrs-shared/distributed_discovery_gpu_test.go','encoding/liftrs-shared/distributed_executor_cuda.go','encoding/liftrs-shared/distributed_pipeline_gpu_test.go','encoding/liftrs-shared/distributed_operands.go','encoding/liftrs-shared/distributed_square.go','encoding/liftrs-shared/distributed_prefix_input.go','encoding/liftrs-shared/distributed_recovery.go','encoding/liftrs-shared/distributed_reference_test.go','encoding/liftrs-shared/distributed_vectors_gpu_test.go','encoding/liftrs-shared/root_shards.go','encoding/liftrs-shared/pipeline_roots_gpu_test.go','encoding/liftrs-cuda/rank_storage.cuh','encoding/liftrs-cuda/preparation.cu','encoding/liftrs-shared/distributed_taylor.go')


def prefixes(target):
    known=1;out=[]
    while known<target:
        size=min(2*known,target);out.append((known,size));known=size
    return out


def project(s,ranks=16,k=None):
    assert 16<=s<=20 and ranks>0 and ranks&(ranks-1)==0
    n=1<<(2*s);width=4 if s<=16 else 8;assert n%ranks==0
    words=n//ranks;base=4*words;ext=width*words
    minimum=1
    while 3*n*width//minimum>112<<30:minimum*=2
    row=dict(s=s,n=n,k=k,word_bytes=width,illustrative_workers=ranks,one_worker_per_gpu=True,
        device_three_arrays_bytes_per_gpu=3*ext,device_three_arrays_bytes_total=3*n*width,payload_only_min_power2_gpus_at112GiB=minimum,
        host_base_array_bytes_per_worker=base,host_extension_array_bytes_per_worker=ext,host_union_bitmap_bytes_per_worker=8*((words+63)//64),
        discovery_base_grid_plus_bitmap_bytes_per_worker=base+8*((words+63)//64),
        newton_root_clone_bytes_per_step_per_worker=base,newton_input_arrays_before_upload_dense_bytes_per_worker=4*base,newton_input_arrays_before_upload_prefix_bytes_per_worker=3*base,
        debug_base_snapshot_bytes_total=12*n,debug_spectrum_bytes_total=width*n,
        taylor_capacity_words=3*(1<<s),taylor_current_capacity_limit_words=1<<20,
        backend_degree_admitted=s<=18,full_distributed_pipeline_run=False,
        not_a_peak_bound='Input-array inventory excludes transfer/native/runtime/GC,other aliases and output allocation;device payload excludes tables/plans/windows/context. Distinct phases must not be summed.')
    if k is not None:
        assert 0<k<n//2
        steps=prefixes(n-k)
        dense_words=2*n*len(steps)
        active_words=sum(min(k+1,size)+min(2*known-1,size)for known,size in steps)
        row.update(newton_steps=len(steps),newton_dense_input_upload_bytes_total=dense_words*width,
            main_fft_count=3*len(steps)+4,
            main_transpose_proof_rows_per_worker=(3*len(steps)+4)*2*2*(ranks-1)*((n//(ranks*ranks)+4095)//4096),
            newton_prefix_input_upload_bytes_total=active_words*width,checkpoint_padded_payload_bytes=4*n*(len(steps)+1),checkpoint_prefix_payload_bytes=4*(k+1+sum(size for _,size in steps)),
            canonical_code_payload_bytes=4*(k+1)+width*n,generator_payload_bytes=4*(n-k),support_payload_bytes=(4 if s<=16 else 8)*k,
            restart_support_bytes_rehashed_per_worker_two_prefix_loads=2*(4 if s<=16 else 8)*k)
    return row


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--output',type=Path,required=True);a=p.parse_args()
    result=dict(status='source-derived projections;not measured memory/runtime or resource admission',revision=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),source_sha256={name:hashlib.sha256((ROOT/name).read_bytes()).hexdigest()for name in SOURCES},rows=[project(s,k=KNOWN_K.get(s))for s in range(16,21)],assumptions=['One worker/GPU,16illustrative workers;exact physical-node packing and resources unadmitted.','Known K from retained exact class-count76647;t5,m2;no measured K19/20.','Three-array minimum ignores extra memory;112GiB/card is a scenario,not an approved allocation.','Newton traffic counts only input uploads;FFTs/resets/window/exchange/checkpoint work unchanged.','Diagnostic serialization buffers,compiler liveness/Go capacity/GC/native memory prevent treating inventories as peak RSS.','Rank-local good/support/root counts and largest subtree domains remain unknown at fulls18.','Distributed pilot remains bounded to s4–8/four owned ranks and native references;production entry and budgets remain separate.'])
    a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))


if __name__=='__main__':main()
