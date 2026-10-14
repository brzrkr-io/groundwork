#!/usr/bin/env bash
set -euo pipefail

# Task 2: print how many lines of the log file $1 are ERROR or FATAL: just
# the number. When there are none, print 0 and still finish with status 0.
# Usage:   bash 2_count_problems.sh LOG
# Example: bash 2_count_problems.sh testdata/app.log
#          6
#          bash 2_count_problems.sh testdata/quiet.log
#          0

# TODO(2): count the ERROR or FATAL lines with one grep.
