#!/bin/bash
# CPU baseline; no GPU is requested by this job.
#SBATCH --job-name=depir-s15
#SBATCH --partition=gpuq
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem=64G
#SBATCH --time=01:00:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0
python3 experiments/hpc-cached-s15/run.py "$1"
