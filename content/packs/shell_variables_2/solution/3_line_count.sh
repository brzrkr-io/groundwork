#!/usr/bin/env bash
set -euo pipefail

# Task 3: count the lines in a file, even when its name has a space.
# Usage:   bash 3_line_count.sh FILE
# Example: bash 3_line_count.sh "testdata/release notes.txt"
#          testdata/release notes.txt: 3 lines

# Quoted, "$1" stays one word even when the name has spaces.
lines=$(wc -l < "$1")
echo "$1: $lines lines"
