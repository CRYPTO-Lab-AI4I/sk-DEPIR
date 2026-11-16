#!/usr/bin/env python3
"""Shared bounded controller: server regressions plus resident client encoding."""
from pathlib import Path
import runpy

HERE = Path(__file__).resolve().parent
if __name__ == '__main__':
    workflow = runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))
    suite = runpy.run_path(str(HERE/'suite.py'))['run_suite']
    workflow['main'](here=HERE, extra_suite=suite)
