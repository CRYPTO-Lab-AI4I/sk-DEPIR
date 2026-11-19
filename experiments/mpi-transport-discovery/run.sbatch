#!/bin/bash
#SBATCH --job-name=depir-mpi-select
#SBATCH --partition=gpuq
#SBATCH --nodes=2
#SBATCH --ntasks-per-node=2
#SBATCH --cpus-per-task=1
#SBATCH --mem=4G
#SBATCH --exclude=gnode03
#SBATCH --time=00:15:00
set -euo pipefail
module load slurm/slurm/25.05 gcc/14.2.0 openmpi/gcc/64/4.1.8
exec python3 -B experiments/mpi-transport-discovery/run.py "$@"
