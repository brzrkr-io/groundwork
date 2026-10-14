## Task 1: Report health with an exit status

You don't need `exit` here. If the last command in the script is a quiet
search, the script's status is the search's status.

### Closer

`grep -q 'PATTERN' "$1"` exits 0 when it finds the pattern and 1 when it
doesn't. The pattern is the exact line text.

### Answer

```bash
grep -q 'status: ok' "$1"
```

## Task 2: Approve or block the deploy

A is the same quiet search as in task 1. B runs when A succeeds, and C runs
when it fails.

### Closer

`A && echo "..." || echo "..."` on one line, with the search from task 1 as
A.

### Answer

```bash
grep -q 'status: ok' "$1" && echo "deploy approved" || echo "deploy blocked"
```

## Task 3: Stop with your own exit code

The check goes before the existing `echo`, so a failing check stops the
script before it prints `preflight ok`. On failure, two things happen: a
message to stderr, then `exit 3`.

### Closer

`grep -q '^version:' "$1" || { ...; }`, with an `echo ... >&2` and an
`exit 3` inside the braces.

### Answer

```bash
grep -q '^version:' "$1" || {
	echo "preflight: no version in $1" >&2
	exit 3
}
```
