#!/usr/bin/env bash
# A pretend deploy for this exercise: results go to stdout, progress and
# warnings go to stderr.
echo "checking api..." >&2
echo "api: ok"
echo "warning: worker restarted once" >&2
echo "worker: ok"
