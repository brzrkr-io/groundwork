#!/usr/bin/env bash
set -euo pipefail

# Task 4: print the name of a backup archive: SERVICE_DATE.tar.gz
# Usage:   bash 4_backup_name.sh SERVICE DATE
# Example: bash 4_backup_name.sh api 2026-10-13
#          api_2026-10-13.tar.gz

service="$1"
date="$2"
# Braces end the name: without them bash reads $service_ as one variable.
echo "${service}_$date.tar.gz"
