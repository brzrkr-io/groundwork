#!/usr/bin/env bash
set -euo pipefail

# Task 1: check the log before reading it.
# Usage:   bash 1_check_input.sh LOG
# No LOG (or an empty one): print "usage: 1_check_input.sh LOG" to stderr,
#                           exit 2.
# LOG is not a regular file: print "cannot read LOG" to stderr, exit 1.
# Otherwise:                 print "LOG: N lines".
# Example: bash 1_check_input.sh testdata/quiet.log
#          testdata/quiet.log: 6 lines

log="${1:-}"

if [[ -z "$log" ]]; then
	echo "usage: 1_check_input.sh LOG" >&2
	exit 2
fi

if [[ ! -f "$log" ]]; then
	echo "cannot read $log" >&2
	exit 1
fi

lines=$(wc -l < "$log")
echo "$log: $lines lines"
