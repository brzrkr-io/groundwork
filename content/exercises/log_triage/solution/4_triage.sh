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

section() {
	echo
	echo "== $1 =="
}

# A failed check stops the report here, thanks to set -e.
bash 1_check_input.sh "$1"

section "levels"
bash 2_levels.sh "$1"

section "failing services"
bash 3_failing_services.sh "$1"

section "slowest requests"
awk '{print $4 " ms " $3}' "$1" | sort -rn | head -n 3
