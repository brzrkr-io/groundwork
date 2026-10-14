#!/usr/bin/env bash
set -euo pipefail

# Task 4: how long is each log?
# Usage:   bash 4_log_lines.sh DIR
# For every file in DIR whose name ends in .log, in alphabetical order,
# print "NAME: N lines", where NAME is the file name without the directory.
# Example: bash 4_log_lines.sh testdata/logs_old
#          billing.log: 3 lines
#          search.log: 2 lines

# Bash expands the pattern to the matching paths, sorted.
for file in "$1"/*.log; do
	lines=$(wc -l < "$file")
	echo "$(basename "$file"): $lines lines"
done
