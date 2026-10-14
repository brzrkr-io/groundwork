#!/usr/bin/env bash
set -euo pipefail

# Task 3: a preflight check with its own exit code.
# Usage:   bash 3_preflight.sh FILE
# When no line of FILE starts with "version:", print
# "preflight: no version in FILE" to stderr and exit 3.
# Otherwise print "preflight ok: FILE" (to stdout) and exit 0.
# Example: bash 3_preflight.sh testdata/health_ok.txt
#          preflight ok: testdata/health_ok.txt

# TODO(3): stop with the error and exit status 3 when there is no version.

echo "preflight ok: $1"
