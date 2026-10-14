#!/usr/bin/env bash
# Task 1: List the commits only on the branch
#
# Print the subject of every commit that is on feature/retries but not
# on main, newest first, one per line.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../1_list_branch_commits.sh
set -euo pipefail

git log --format=%s main..feature/retries
