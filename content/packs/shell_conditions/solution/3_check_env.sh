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

if [[ -z "$env" ]]; then
	echo "usage: 3_check_env.sh ENV" >&2
	exit 2
fi

if [[ "$env" == "production" || "$env" == "staging" ]]; then
	echo "$env is a valid environment"
else
	echo "unknown environment: $env" >&2
	exit 2
fi
