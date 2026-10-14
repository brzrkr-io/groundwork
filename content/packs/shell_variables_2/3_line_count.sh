#!/usr/bin/env bash
set -euo pipefail

# Task 3: count the lines in a file, even when its name has a space.
# Usage:   bash 3_line_count.sh FILE
# Example: bash 3_line_count.sh "testdata/release notes.txt"
#          testdata/release notes.txt: 3 lines

# TODO(3): this fails with "ambiguous redirect" when the file name has a
# space. Fix it.
lines=$(wc -l < $1)
echo "$1: $lines lines"
