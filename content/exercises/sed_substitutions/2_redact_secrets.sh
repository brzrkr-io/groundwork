#!/usr/bin/env bash
set -euo pipefail

# Task 2: print the env file given as $1 with every secret hidden. A secret
# is a line whose key ends in _PASSWORD or _TOKEN: print its key, then =,
# then REDACTED. Print every other line unchanged. For example:
#   bash 2_redact_secrets.sh testdata/app.env
# prints
#   DB_PASSWORD=REDACTED
# where the file says DB_PASSWORD=s3cr3t-Pa55.

# TODO(2): print the file with the secret values replaced, using sed.
