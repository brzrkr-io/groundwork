#!/usr/bin/env bash
set -euo pipefail

# Task 2: a deploy gate.
# Usage:   bash 2_gate.sh FILE
# Print "deploy approved" when FILE has the line "status: ok",
# otherwise print "deploy blocked". Exit 0 either way.
# Example: bash 2_gate.sh testdata/health_down.txt
#          deploy blocked

grep -q 'status: ok' "$1" && echo "deploy approved" || echo "deploy blocked"
