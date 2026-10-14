#!/usr/bin/env bash
set -euo pipefail

# Task 3: add up request counts. Each line of FILE is "SERVICE COUNT".
# Usage:   bash 3_total_requests.sh FILE
# Print "total: N", where N is the sum of the counts.
# Example: bash 3_total_requests.sh testdata/requests_night.txt
#          total: 350

total=0

# read splits each line at spaces: the service name into _ (not needed),
# the rest into count.
while read -r _ count; do
	total=$((total + count))
done < "$1"

echo "total: $total"
