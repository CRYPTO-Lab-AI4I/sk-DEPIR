#!/usr/bin/env python3
"""Repeat the failed tree boundary with fingerprints and durable mismatch capture."""
from pathlib import Path
import runpy
import sys
HERE=Path(__file__).resolve().parent
TREE=runpy.run_path(str(HERE.parent/'root-merge-tree/run.py'))


def suite(context):
    context['env'].update(PIR_ROOT_TREE_DIAGNOSTICS='1',PIR_ROOT_TREE_REPEATS='12')
    return TREE['suite'](context)


if __name__=='__main__':
    if '--check'in sys.argv:print('same2node/2H200/4logical rank/4CPU/8GiBhost/15min envelope;12 largest-case trials; diagnostic proofs/capture, no staging workaround')
    else:TREE['BASE']['main'](here=HERE,extra_suite=suite)
