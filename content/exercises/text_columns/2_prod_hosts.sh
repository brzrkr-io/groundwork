#!/usr/bin/env bash
set -euo pipefail

# Task 2: print the name of every host in the CSV file $1 whose env is
# exactly prod (preprod doesn't count), one per line, in file order.
# Usage:   bash 2_prod_hosts.sh CSV
# Example: bash 2_prod_hosts.sh testdata/hosts_eu.csv
#          eu-web-01
#          eu-db-01
#          eu-db-02

# TODO(2): print the names of the prod hosts with awk.
