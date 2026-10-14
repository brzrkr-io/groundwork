#!/usr/bin/env bash
set -euo pipefail

# Task 3: for each service with ERROR lines in the log file $1, print the
# number of ERROR lines and the service name, highest count first. The
# tests ignore spaces at the start of each line.
# Usage:   bash 3_errors_per_service.sh LOG
# Example: bash 3_errors_per_service.sh testdata/app.log
#                4 payments
#                2 api
#                1 checkout

grep ERROR "$1" | cut -d' ' -f3 | sort | uniq -c | sort -rn
