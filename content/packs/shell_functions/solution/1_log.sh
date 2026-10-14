#!/usr/bin/env bash
set -euo pipefail

# Task 1: two log helpers for a deploy script.
# Usage:   bash 1_log.sh SERVICE
# log_info MESSAGE  prints "[INFO] MESSAGE" to stdout.
# log_error MESSAGE prints "[ERROR] MESSAGE" to stderr.

log_info() {
	echo "[INFO] $1"
}

log_error() {
	echo "[ERROR] $1" >&2
}

log_info "deploy started: $1"
log_info "pulling image"
log_error "image pull slow, retrying"
log_info "deploy finished: $1"
