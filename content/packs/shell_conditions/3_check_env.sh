#!/usr/bin/env bash
set -euo pipefail

# Task 3: validate a deploy environment.
# Usage:   bash 3_check_env.sh ENV
# No ENV (or an empty one): print "usage: 3_check_env.sh ENV" to stderr,
#                           exit 2.
# production or staging:    print "ENV is a valid environment".
# anything else:            print "unknown environment: ENV" to stderr,
#                           exit 2.
# Example: bash 3_check_env.sh staging
#          staging is a valid environment

env="${1:-}"

# TODO(3): check for a missing ENV first, then check its value.
