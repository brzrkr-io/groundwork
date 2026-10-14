## Task 1: Print the error lines

grep prints the lines that contain a pattern. The pattern here is the level
itself, and the file is the script's first argument.

### Closer

`grep PATTERN FILE`, with `"$1"` as the file (in double quotes, as always
for arguments): `grep ... "$1"`.

### Answer

```bash
grep ERROR "$1"
```

## Task 2: Count the problems

You need two flags: one that prints a count instead of lines, and one that
lets `|` mean "or". Then think about what happens when the count is 0.

### Closer

`grep -cE 'A|B' "$1"` counts the lines with A or B. On `testdata/quiet.log`
it prints `0` and exits with status 1, and `set -e` stops the script there.
Add `|| true` after it.

### Answer

```bash
grep -cE 'ERROR|FATAL' "$1" || true
```

`|| true` turns grep's "no match" status into success; the `0` it printed
stays.

## Task 3: Hide the health checks

Every health-check line contains `/healthz`. You want all the *other*
lines, so look for the flag that inverts the match.

### Closer

`-v` prints the lines that don't match: `grep -v PATTERN "$1"`.

### Answer

```bash
grep -v /healthz "$1"
```

## Task 4: List the failed request ids

Three steps joined with pipes. First keep the lines with `ERROR` followed by
the service, then cut each line down to `req=...`, then drop the `req=`.

### Closer

- Step 1: the pattern is `ERROR`, a space and the service, in **double**
  quotes so `$2` is swapped in: `grep "ERROR ..." "$1"`. A space after the
  service stops `api` from matching a longer name such as `api-gateway`.
- Step 2: `grep -o` with `req=` followed by a run of lowercase letters and
  digits (the bracket pattern from the Learn section).
- Step 3: `cut` with `=` as the delimiter, keeping field 2.

Run it after each step and look at the output before adding the next.

### Answer

```bash
grep "ERROR $2 " "$1" | grep -o 'req=[a-z0-9]*' | cut -d= -f2
```
