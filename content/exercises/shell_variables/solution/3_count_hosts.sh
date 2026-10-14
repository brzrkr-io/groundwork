#!/usr/bin/env bash
set -euo pipefail

# Task 3: count the hosts in an inventory file, one host per line.
# Usage:   bash 3_count_hosts.sh FILE
# Example: bash 3_count_hosts.sh testdata/web_hosts.txt
#          testdata/web_hosts.txt: 4 hosts

# "< file" feeds the file to wc, so it prints only the number, not the name.
count=$(wc -l < "$1")
echo "$1: $count hosts"
