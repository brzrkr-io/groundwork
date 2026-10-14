#!/usr/bin/env bash
# Task 1: Find the commits that touched a value
#
# Print every commit that added or removed the text "max_connections: 500"
# (anywhere in the repo), newest first, one per line, as:
# <author name>: <subject>
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_find_connection_limit.sh
set -euo pipefail

git log -S 'max_connections: 500' --format='%an: %s'
