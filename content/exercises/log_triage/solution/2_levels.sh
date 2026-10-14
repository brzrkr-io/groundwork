#!/usr/bin/env bash
set -euo pipefail

# Task 2: count the lines of each level, in this order: ERROR, WARN, INFO.
# Usage:   bash 2_levels.sh LOG
# Print "LEVEL: N" for each level, even when N is 0.
# Example: bash 2_levels.sh testdata/quiet.log
#          ERROR: 0
#          WARN: 1
#          INFO: 5

for level in ERROR WARN INFO; do
	# The spaces around the level match the level column only. grep -c
	# still prints 0 when nothing matches, but exits 1: || true keeps going.
	count=$(grep -c " $level " "$1" || true)
	echo "$level: $count"
done
