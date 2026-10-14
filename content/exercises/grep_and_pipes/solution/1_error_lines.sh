#!/usr/bin/env bash
set -euo pipefail

# Task 1: print every ERROR line of the log file given as $1, whole and in
# the original order.
# Usage:   bash 1_error_lines.sh LOG
# Example: bash 1_error_lines.sh testdata/worker.log
#          2026-10-13T22:00:15Z ERROR mailer msg="smtp auth failed" req=c2e8f4
#          2026-10-13T22:00:40Z ERROR worker msg="job failed: reports" req=1d4f7a
#          2026-10-13T22:00:52Z ERROR mailer msg="smtp auth failed" req=f85b21

grep ERROR "$1"
