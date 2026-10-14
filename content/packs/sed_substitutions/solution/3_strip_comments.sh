#!/usr/bin/env bash
set -euo pipefail

# Task 3: print the env file given as $1 without its comment lines (lines
# that start with #) and without its empty lines. Every other line is
# printed unchanged. For example:
#   bash 3_strip_comments.sh testdata/worker.env
# prints
#   WORKER_CONCURRENCY=8
#   LOG_LEVEL=info
#   QUEUE_URL=amqp://mq.internal:5672/jobs
#   QUEUE_PASSWORD=rabbit-Ears-77
#   METRICS_TOKEN=bXQtOWY4ZTdk==
#   METRICS_PORT=9100

sed -e '/^#/d' -e '/^$/d' "$1"
