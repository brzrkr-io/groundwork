#!/usr/bin/env bash
# A pretend deploy check: it prints the two settings it reads from its
# environment, or <unset> when one isn't there.
echo "DEPLOY_ENV=${DEPLOY_ENV-<unset>}"
echo "LOG_LEVEL=${LOG_LEVEL-<unset>}"
