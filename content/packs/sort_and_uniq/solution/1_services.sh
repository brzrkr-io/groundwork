#!/usr/bin/env bash
set -euo pipefail

# Task 1: print every service name (the third column) in the log file $1,
# each once, sorted A to Z.
# Usage:   bash 1_services.sh LOG
# Example: bash 1_services.sh testdata/worker.log
#          exporter
#          mailer
#          scheduler
#          worker

cut -d' ' -f3 "$1" | sort -u
