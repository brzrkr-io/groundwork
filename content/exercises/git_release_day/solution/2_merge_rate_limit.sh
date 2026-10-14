#!/usr/bin/env bash
# Task 2: Merge the rate-limit feature
#
# Merge feature/rate-limit into main with a merge commit and the default
# message, then delete the feature branch.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../2_merge_rate_limit.sh
set -euo pipefail

git merge --no-edit feature/rate-limit
git branch -d feature/rate-limit
