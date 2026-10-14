#!/usr/bin/env bash
set -euo pipefail

# Task 2: stop the process whose PID is $1. Send TERM, then check every 0.1
# seconds, at most 10 times, whether it's gone. If it's still there, send
# KILL. Then print which signal stopped it, exactly one of:
#   stopped by TERM (exit code 143)
#   killed by KILL (exit code 137)
# For example, after bash setup.sh:
#   bash 2_stop.sh "$(pgrep -x report-export)"
# prints
#   killed by KILL (exit code 137)

pid="$1"
kill -TERM "$pid"
for _ in $(seq 10); do
	sleep 0.1
	if ! kill -0 "$pid" 2> /dev/null; then
		echo "stopped by TERM (exit code 143)"
		exit 0
	fi
done
kill -KILL "$pid"
echo "killed by KILL (exit code 137)"
