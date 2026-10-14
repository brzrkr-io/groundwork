#!/usr/bin/env bash
set -euo pipefail

# Task 4: check that the tools a script needs are installed.
# Usage:   bash 4_require.sh TOOL...
# require TOOL prints "missing tool: TOOL" to stderr and stops the whole
# script with exit status 1 when TOOL isn't installed.
# Example: bash 4_require.sh grep sed
#          all tools found

require() {
	if ! command -v "$1" > /dev/null; then
		echo "missing tool: $1" >&2
		exit 1
	fi
}

for tool in "$@"; do
	require "$tool"
done
echo "all tools found"
