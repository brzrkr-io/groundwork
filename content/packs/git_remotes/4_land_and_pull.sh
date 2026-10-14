#!/usr/bin/env bash
# Task 4: Land the fix and pull it
#
# In web, switch to main, fast-forward it to fix/timeout (refusing any
# other kind of merge), and push main to origin. Then play Alex: in alex,
# pull main, again refusing anything but a fast-forward.
#
# This script runs inside the sandbox folder (where origin.git and alex/
# are): cd sandbox && bash ../4_land_and_pull.sh
set -euo pipefail

# TODO(4): in web, fast-forward main to fix/timeout and push it; then pull in alex.
