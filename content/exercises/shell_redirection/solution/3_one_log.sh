#!/usr/bin/env bash
set -euo pipefail

# Task 3: Put everything in one log
#
# Run "bash testdata/deploy.sh" with both stdout and stderr going into
# all.log, in the order they were printed, so nothing reaches the terminal.
#
# Usage:   bash 3_one_log.sh

# > first, then 2>&1: stderr goes wherever stdout points by then.
bash testdata/deploy.sh > all.log 2>&1
