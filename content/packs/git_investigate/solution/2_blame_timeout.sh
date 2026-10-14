#!/usr/bin/env bash
# Task 2: Find who last changed a line
#
# Print the commit that last changed line 4 of config/app.yaml (the
# timeout line) as:
# <author name>: <subject>
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_blame_timeout.sh
set -euo pipefail

commit=$(git blame --porcelain -L 4,4 config/app.yaml | head -n 1 | cut -d ' ' -f 1)
git log -1 --format='%an: %s' "$commit"
