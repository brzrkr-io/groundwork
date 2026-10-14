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

# EOF without quotes, so bash fills in the variables.
cat > app.yaml <<EOF
service: $service
replicas: $replicas
image: registry.example.com/$service:stable
EOF
