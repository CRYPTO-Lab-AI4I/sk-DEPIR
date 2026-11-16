#!/usr/bin/env python3
"""Reuse complete workflow validation, then count exact s16/s17 dimensions on GPUs."""
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent
if __name__=='__main__':
 runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/run.py'))['main'](here=HERE,dimension_default=True)
