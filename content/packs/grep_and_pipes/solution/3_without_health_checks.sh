#!/usr/bin/env bash
set -euo pipefail

# Task 3: print the log file $1 without the health-check lines: every line
# that does not contain /healthz, in the original order.
# Usage:   bash 3_without_health_checks.sh LOG
# Example: bash 3_without_health_checks.sh testdata/quiet.log
#          2026-10-14T03:00:04Z INFO auth msg="token refreshed" req=7b0f3d
#          2026-10-14T03:00:13Z WARN checkout msg="cache miss rate high" req=58e2c7
#          2026-10-14T03:00:26Z INFO payments msg="charge settled" req=0e9b84

grep -v /healthz "$1"
