#!/usr/bin/env bash
# Task 3: Fetch as Alex
#
# Play Alex: in alex, fetch from origin, then print the subject of each
# commit on origin/fix/timeout that Alex's main doesn't have, newest first,
# one per line. Don't change Alex's branches or files.
#
# This script runs inside the sandbox folder (where origin.git and alex/
# are): cd sandbox && bash ../3_review_as_alex.sh
set -euo pipefail

cd alex
# fetch updates origin/* only; Alex's main and files stay as they are.
git fetch
git log --format=%s main..origin/fix/timeout
