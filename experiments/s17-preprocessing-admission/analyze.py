#!/usr/bin/env python3
"""Reconcile admission counts and the common complete-workflow regression suite."""
from pathlib import Path
import runpy
HERE=Path(__file__).resolve().parent
if __name__=='__main__':
 runpy.run_path(str(HERE.parent/'gpu-preprocessing-workflow/analyze.py'))['main'](here=HERE,require_dimension=True)
