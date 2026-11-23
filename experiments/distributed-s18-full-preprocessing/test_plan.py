#!/usr/bin/env python3
"""No-GPU configuration and launcher checks before requesting the allocation."""
from pathlib import Path
import copy
import json
import runpy
import subprocess
import tempfile

HERE=Path(__file__).resolve().parent
M=runpy.run_path(str(HERE/'run.py'))
b=json.loads((HERE/'budget.json').read_text()); M['validate'](b)
for field,value in [('rss_GiB_per_rank',300),('host_request_GiB_per_node',800),
                    ('host_reserve_GiB_per_rank',300),('owned_gpu_GiB_per_rank',96),
                    ('controller_hours',101),('retained_GiB',2048)]:
    bad=copy.deepcopy(b);bad[field]=value
    try: M['validate'](bad)
    except AssertionError: pass
    else: raise AssertionError('accepted unsafe '+field)
# Scalar policy is also read by the existing actual MPI command, which --check
# validates without MPI/CUDA initialization. No independent GPU rebuild needed.
binary=HERE.parents[1]/'.cache/liftrs/distributed-discovery-restart/recovery-01/pir-preprocess-mpi'
with tempfile.TemporaryDirectory() as directory:
    directory=Path(directory)
    for s in (8,18):
        config=M['config'](s,directory/'output',directory/'library.so',b)
        path=directory/'config.json'; path.write_text(json.dumps(config))
        result=subprocess.run([str(binary),'--config',str(path),'--check-config'],capture_output=True,text=True,check=True,timeout=5)
        assert json.loads(result.stdout)==config
    # The command's coefficient count is roots+1, so the largest allowed list
    # must leave that final coefficient within the admitted full shard.
    assert config['max_root_words']+1==config['max_local_words']==1<<32
script=(HERE/'run.sbatch').read_text()
assert all(text in script for text in ['--nodes=8','--ntasks-per-node=2','--cpus-per-task=4','--mem=864G','--gres=gpu:h200:2','--time=4-04:00:00'])
# A full launch remains blocked even when a valid fresh name is supplied.
result=subprocess.run(['python3','-B',str(HERE/'run.py'),'not-admitted'],capture_output=True,text=True,timeout=5)
assert result.returncode and 'resource agreement has not been recorded' in result.stderr
print(json.dumps(dict(status='passed',unsafe_budgets_rejected=6,actual_command_configs=[8,18],admission_block=True)))
