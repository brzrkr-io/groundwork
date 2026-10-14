#!/usr/bin/env bash
set -euo pipefail

# Task 1: print the name and IP of every host in the CSV file $1 (header:
# name,ip,env,role) as name,ip, without the header line.
# Usage:   bash 1_names_and_ips.sh CSV
# Example: bash 1_names_and_ips.sh testdata/hosts_eu.csv
#          eu-web-01,10.20.1.11
#          eu-web-02,10.20.2.12
#          eu-db-01,10.20.1.21
#          eu-queue-01,10.20.3.31
#          eu-db-02,10.20.1.22

# TODO(1): skip the header, then keep the name and IP columns with cut.
