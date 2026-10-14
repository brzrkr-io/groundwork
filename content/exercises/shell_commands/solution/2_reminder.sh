#!/usr/bin/env bash
set -euo pipefail

# Task 2: print exactly these two lines, dollar sign and quotes included:
#   Set $IMAGE_TAG before you deploy.
#   Then run "make deploy".

# Single quotes: bash leaves everything inside alone, $ and " included.
echo 'Set $IMAGE_TAG before you deploy.'
echo 'Then run "make deploy".'
