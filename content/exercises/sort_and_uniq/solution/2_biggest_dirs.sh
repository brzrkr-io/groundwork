#!/usr/bin/env bash
set -euo pipefail

# Task 2: print the 5 biggest lines of the disk usage file $1 (size in
# kilobytes, a tab, a directory), biggest first, each line unchanged.
# Usage:   bash 2_biggest_dirs.sh FILE
# Example: bash 2_biggest_dirs.sh testdata/du_web01.txt
#          30412   /opt/app/releases
#          12804   /var/lib/docker
#          2310    /var/log/journal
#          1024    /var/log/app
#          880     /var/cache/apt

sort -rn "$1" | head -n 5
