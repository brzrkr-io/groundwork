#!/usr/bin/env bash
set -euo pipefail

# Task 3: print STATUS PATH (one space between) for every request in the
# nginx access log $1 with a status of 500 or more, in file order.
# Usage:   bash 3_server_errors.sh LOG
# Example: bash 3_server_errors.sh testdata/access.log
#          502 /api/checkout
#          500 /api/orders
#          503 /api/search?q=boots
#          504 /api/orders

awk '$9 >= 500 {print $9, $7}' "$1"
