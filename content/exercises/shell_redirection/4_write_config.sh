#!/usr/bin/env bash
set -euo pipefail

# Task 4: Write the config with a here-doc
#
# Write app.yaml with a here-doc. SERVICE is $1 and REPLICAS is $2:
#   service: SERVICE
#   replicas: REPLICAS
#   image: registry.example.com/SERVICE:stable
#
# Usage:   bash 4_write_config.sh SERVICE REPLICAS
# Example: bash 4_write_config.sh api 3

service="$1"
replicas="$2"

# TODO(4): write the three lines of app.yaml with cat and a here-doc.
