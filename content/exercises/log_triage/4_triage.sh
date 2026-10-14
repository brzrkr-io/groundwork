#!/usr/bin/env bash
set -euo pipefail

# Task 4: the full triage report, built from your other three scripts.
# Usage:   bash 4_triage.sh LOG
# Example: bash 4_triage.sh testdata/quiet.log
#          testdata/quiet.log: 6 lines
#
#          == levels ==
#          ERROR: 0
#          WARN: 1
#          INFO: 5
#
#          == failing services ==
#
#          == slowest requests ==
#          3100 ms reports
#          2400 ms reports
#          870 ms billing

# TODO(4): define a section function, then print the report.
