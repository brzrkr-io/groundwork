#!/usr/bin/env bash
# Task 2: Find who last changed a line
#
# Print the commit that last changed line 4 of config/app.yaml (the
# timeout line) as:
# <author name>: <subject>
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_blame_timeout.sh
set -euo pipefail

# TODO(2): find the commit with git blame, then print it as "author: subject".
