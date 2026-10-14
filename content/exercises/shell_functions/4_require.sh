#!/usr/bin/env bash
set -euo pipefail

# Task 4: check that the tools a script needs are installed.
# Usage:   bash 4_require.sh TOOL...
# require TOOL prints "missing tool: TOOL" to stderr and stops the whole
# script with exit status 1 when TOOL isn't installed.
# Example: bash 4_require.sh grep sed
#          all tools found

# TODO(4): define require here.

for tool in "$@"; do
	require "$tool"
done
echo "all tools found"
