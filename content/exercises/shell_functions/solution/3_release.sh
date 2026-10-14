#!/usr/bin/env bash
set -euo pipefail

# Task 3: plan a release. Each line of FILE is "NAME TAG".
# Usage:   bash 3_release.sh FILE
# image_ref NAME TAG prints "registry.example.com/NAME:TAG".
# Example: bash 3_release.sh testdata/hotfix.txt
#          deploy payments -> registry.example.com/payments:2.0.2

# local keeps name and tag inside the function, so it can't overwrite the
# loop's variables of the same names.
image_ref() {
	local name="$1"
	local tag="$2"
	echo "registry.example.com/$name:$tag"
}

while read -r name tag; do
	ref=$(image_ref "$name" "$tag")
	echo "deploy $name -> $ref"
done < "$1"
