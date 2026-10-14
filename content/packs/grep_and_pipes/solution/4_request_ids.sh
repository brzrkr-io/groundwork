#!/usr/bin/env bash
set -euo pipefail

# Task 4: print the request ids (the part after req=) of the ERROR lines of
# service $2 in the log file $1: one id per line, in the original order.
# Usage:   bash 4_request_ids.sh LOG SERVICE
# Example: bash 4_request_ids.sh testdata/app.log payments
#          7f3a9c
#          0f1e2d
#          d4e5f6

grep "ERROR $2 " "$1" | grep -o 'req=[a-z0-9]*' | cut -d= -f2
