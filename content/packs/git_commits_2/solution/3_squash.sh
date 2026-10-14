#!/usr/bin/env bash
# Task 3: Squash the last two commits
#
# "Add health check script" and "Fix the health check timeout" belong
# together. Replace them with one commit with the message:
# Add health check with a 5s timeout
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_squash.sh
set -euo pipefail

git reset --soft HEAD~2
git commit -m "Add health check with a 5s timeout"
