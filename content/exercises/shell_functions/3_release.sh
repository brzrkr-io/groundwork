#!/usr/bin/env bash
set -euo pipefail

# Task 3: plan a release. Each line of FILE is "NAME TAG".
# Usage:   bash 3_release.sh FILE
# image_ref NAME TAG prints "registry.example.com/NAME:TAG".
# Example: bash 3_release.sh testdata/hotfix.txt
#          deploy payments -> registry.example.com/payments:2.0.2

# TODO(3): define image_ref here, with local variables for its arguments.

while read -r name tag; do
	ref=$(image_ref "$name" "$tag")
	echo "deploy $name -> $ref"
done < "$1"
