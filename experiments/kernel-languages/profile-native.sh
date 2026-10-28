#!/bin/sh
# Sample only the caller-selected benchmark PID. Run outside a sandbox that
# forbids Mach task inspection; output is diagnostic, not a timing replacement.
set -eu
if [ "$#" -ne 2 ]; then
  echo 'usage: profile-native.sh BENCHMARK_PID OUTPUT_FILE' >&2
  exit 2
fi
exec /usr/bin/sample "$1" 20 10 -mayDie -file "$2"
