#!/usr/bin/env bash
set -euo pipefail

# Task 1: answer "is this service healthy?" with the exit status alone.
# Usage:   bash 1_is_healthy.sh FILE
# Exit 0 when FILE has the line "status: ok", exit 1 otherwise.
# Print nothing.
# Example: bash 1_is_healthy.sh testdata/health_ok.txt; echo $?
#          0

# TODO(1): search FILE quietly for the line "status: ok".
