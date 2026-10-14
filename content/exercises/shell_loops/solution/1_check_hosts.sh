#!/usr/bin/env bash
set -euo pipefail

# Task 1: go through every host given as an argument.
# Usage:   bash 1_check_hosts.sh HOST...
# Print "checking HOST" for each host, in order, then "hosts checked: N".
# Example: bash 1_check_hosts.sh web-01 db-01
#          checking web-01
#          checking db-01
#          hosts checked: 2

for host in "$@"; do
	echo "checking $host"
done
echo "hosts checked: $#"
