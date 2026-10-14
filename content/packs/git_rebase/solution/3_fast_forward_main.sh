#!/usr/bin/env bash
# Task 3: Fast-forward main and clean up
#
# Switch to main, bring in feature/retries as a fast-forward only (no
# merge commit), then delete the feature branch.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_fast_forward_main.sh
set -euo pipefail

git switch main
git merge --ff-only feature/retries
git branch -d feature/retries
