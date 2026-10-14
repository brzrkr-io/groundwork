#!/usr/bin/env bash
set -euo pipefail

# Task 2: check port numbers.
# Usage:   bash 2_ports.sh PORT...
# valid_port PORT returns status 0 when PORT is from 1 to 65535, and 1
# otherwise. Every PORT is a whole number.
# Example: bash 2_ports.sh 8080 0
#          8080 ok
#          0 invalid

# TODO(2): define valid_port here.

for port in "$@"; do
	if valid_port "$port"; then
		echo "$port ok"
	else
		echo "$port invalid"
	fi
done
