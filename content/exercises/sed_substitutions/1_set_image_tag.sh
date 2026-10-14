#!/usr/bin/env bash
set -euo pipefail

# Task 1: print testdata/deploy.yaml with the api container's image tag
# replaced by the tag given as $1. Every other line stays as it is,
# including the log-shipper image. For example:
#   bash 1_set_image_tag.sh 1.5.0
# prints the file with this line changed:
#   image: registry.example.com/api:1.4.1
# to:
#   image: registry.example.com/api:1.5.0

# TODO(1): print the file with the new api image tag, using sed.
