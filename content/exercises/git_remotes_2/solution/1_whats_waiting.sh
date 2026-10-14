#!/usr/bin/env bash
# Task 1: See what is waiting on origin
#
# Fetch from origin, then print the subject of each commit on origin/main
# that your main doesn't have, newest first, one per line. Don't change
# your branches.
#
# This script runs inside your clone: cd sandbox/web && bash ../../1_whats_waiting.sh
set -euo pipefail

git fetch
git log --format=%s main..origin/main
