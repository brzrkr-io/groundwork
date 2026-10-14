#!/usr/bin/env bash
set -euo pipefail

# Task 4: print the N client IPs (first column) with the most requests in
# the access log $1, each with its count, most first. N is $2, or 3 when
# there's no $2. The tests ignore spaces at the start of each line.
# Usage:   bash 4_top_ips.sh LOG [N]
# Example: bash 4_top_ips.sh testdata/access.log
#                8 203.0.113.7
#                6 198.51.100.23
#                5 192.0.2.44

# TODO(4): count the requests per IP, most first, and keep the top N.
