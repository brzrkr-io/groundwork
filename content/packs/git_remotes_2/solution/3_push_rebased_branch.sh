#!/usr/bin/env bash
# Task 3: Push the rebased feature branch
#
# Switch to feature/retries, rebase it onto main, then push it to origin
# with the force option that refuses if someone else pushed to it.
#
# This script runs inside your clone: cd sandbox/web && bash ../../3_push_rebased_branch.sh
set -euo pipefail

git switch feature/retries
git rebase main
# The rebase replaced the pushed commits, so a plain push is rejected.
git push --force-with-lease
