#!/usr/bin/env bash
set -euo pipefail

# Task 4: print "total bytes: N", where N is the sum of the response sizes
# (the field after the status) of every request in the nginx access log $1.
# Usage:   bash 4_total_bytes.sh LOG
# Example: bash 4_total_bytes.sh testdata/access.log
#          total bytes: 65684

# TODO(4): add up the response sizes with awk and print the total at the end.
