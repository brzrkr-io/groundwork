#!/usr/bin/env bash
# Task 4: Print the release notes
#
# Print the subject of every commit after v1.3.0 up to and including
# v1.4.0, newest first, skipping merge commits, each as "- <subject>".
#
# This script runs inside the sandbox repo: cd sandbox && bash ../4_release_notes.sh
set -euo pipefail

# TODO(4): print the subjects between the two tags as a list.
