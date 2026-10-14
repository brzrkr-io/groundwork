#!/usr/bin/env bash
set -euo pipefail

# Task 1: check that a config file exists before a deploy.
# Usage:   bash 1_check_config.sh PATH
# When PATH is a regular file, print "found PATH".
# Otherwise print "missing PATH" to stderr and exit 1.
# Example: bash 1_check_config.sh testdata/app.conf
#          found testdata/app.conf

if [[ -f "$1" ]]; then
	echo "found $1"
else
	echo "missing $1" >&2
	exit 1
fi
