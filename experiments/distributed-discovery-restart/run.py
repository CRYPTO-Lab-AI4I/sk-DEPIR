#!/usr/bin/env python3
"""Larger representative cold/discovery-cache/prefix recovery within admitted limits."""
from pathlib import Path
import runpy,sys
HERE=Path(__file__).resolve().parent
HOST=runpy.run_path(str(HERE.parent/'distributed-host-admission/run.py'))
REFERENCE=runpy.run_path(str(HERE/'reference.py'))['REFERENCE'];VECTORS=HOST['VECTORS'];PROOF=HOST['PROOF']
def suite(c):
 c['run']('discovery-restart-review-tests',['python3','-B','experiments/distributed-discovery-restart/test_review.py'],30)
 return HOST['suite'](c,reference=REFERENCE,proof_mode='summary',recoveries=(0,1,8),seed=1603)
def review(rows,directory,known=1,source=None):return HOST['review'](rows,directory,known,source,REFERENCE,'summary')
if __name__=='__main__':
 if '--check'in sys.argv:print('same approved2nodes/2H200/8cores/16GiBhost/8workers;32MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile/720scontroller/15minSlurm;s10 cold+sealed-discovery+root+prefix8 recovery;summary logs explicit;no submission')
 else:VECTORS['BASE']['main'](here=HERE,extra_suite=suite)
