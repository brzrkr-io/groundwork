#!/usr/bin/env bash
set -euo pipefail

# Task 1: is last night's backup there?
# Usage:   bash 1_backup_exists.sh FILE
# When FILE is a regular file, print "backup found: FILE".
# Otherwise print "no backup at FILE" to stderr and exit 1.
# Example: bash 1_backup_exists.sh testdata/backups/db_2026-10-12.tar.gz
#          backup found: testdata/backups/db_2026-10-12.tar.gz

# [[ and ]] are words of their own, so they need spaces around them.
if [[ -f "$1" ]]; then
	echo "backup found: $1"
else
	echo "no backup at $1" >&2
	exit 1
fi
