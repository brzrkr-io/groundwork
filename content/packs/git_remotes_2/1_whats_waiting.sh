#!/usr/bin/env bash
# Task 1: See what's waiting on origin
#
# Fetch from origin, then print the subject of each commit on origin/main
# that your main doesn't have, newest first, one per line. Don't change
# your branches.
#
# This script runs inside your clone: cd sandbox/web && bash ../../1_whats_waiting.sh
set -euo pipefail

# TODO(1): fetch, then print the subjects in main..origin/main.
