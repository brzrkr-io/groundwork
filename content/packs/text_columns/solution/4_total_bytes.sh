#!/usr/bin/env bash
set -euo pipefail

# Task 4: print "total bytes: N", where N is the sum of the response sizes
# (the field after the status) of every request in the nginx access log $1.
# Usage:   bash 4_total_bytes.sh LOG
# Example: bash 4_total_bytes.sh testdata/access.log
#          total bytes: 65684

awk '{total += $10} END {print "total bytes: " total}' "$1"
