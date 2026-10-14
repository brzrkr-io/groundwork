#!/usr/bin/env bash
set -euo pipefail

# Task 3: which services logged the most errors?
# Usage:   bash 3_failing_services.sh LOG
# Print "SERVICE: N" for every service with ERROR lines, most errors first.
# A log without errors prints nothing and exits 0.
# Example: bash 3_failing_services.sh testdata/incident.log
#          payments: 4
#          checkout: 2
#          auth: 1

# awk, unlike grep, succeeds when nothing matches, so pipefail stays happy
# on a quiet log. The last awk turns uniq's "   4 payments" into
# "payments: 4".
awk '$2 == "ERROR" {print $3}' "$1" |
	sort |
	uniq -c |
	sort -rn |
	awk '{print $2 ": " $1}'
