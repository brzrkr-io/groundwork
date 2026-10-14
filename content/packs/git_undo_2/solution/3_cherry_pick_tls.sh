#!/usr/bin/env bash
# Task 3: Copy the rescued commit onto main
#
# Copy the commit at rescue-tls onto main as a new commit on top, then
# delete the rescue-tls branch.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_cherry_pick_tls.sh
set -euo pipefail

git cherry-pick rescue-tls
git branch -D rescue-tls
