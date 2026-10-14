#!/usr/bin/env bash
# Task 3: Bump the version and tag the release
#
# Set VERSION to 1.4.0, commit it with the message "Release v1.4.0",
# and tag that commit v1.4.0.
#
# This script runs inside the sandbox repo: cd sandbox && bash ../3_tag_release.sh
set -euo pipefail

echo "1.4.0" > VERSION
git add VERSION
git commit -m "Release v1.4.0"
git tag v1.4.0
