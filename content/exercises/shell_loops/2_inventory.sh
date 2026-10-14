#!/usr/bin/env bash
set -euo pipefail

# Task 2: turn an inventory file into ssh commands.
# Usage:   bash 2_inventory.sh FILE
# For every line of FILE that is not blank and does not start with "#",
# print "ssh deploy@HOST uptime".
# Example: bash 2_inventory.sh testdata/inventory_staging.txt
#          ssh deploy@web-01.staging.example.com uptime
#          ssh deploy@db-02.staging.example.com uptime

# TODO(2): read FILE line by line, skip blanks and comments, print the rest.
