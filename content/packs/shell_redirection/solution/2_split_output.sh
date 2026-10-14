#!/usr/bin/env bash
set -euo pipefail

# Task 2: Split results and warnings
#
# Run "bash testdata/deploy.sh". Send what it prints to stdout into
# result.txt and what it prints to stderr into warnings.txt, so nothing
# reaches the terminal.
#
# Usage:   bash 2_split_output.sh

bash testdata/deploy.sh > result.txt 2> warnings.txt
