## Task 1: Check the input

Copy `$1` into a variable with an empty default, so `set -u` doesn't stop a
call with no argument. Then two `if`s, each printing to stderr and exiting,
and finally the count.

### Closer

`log="${1:-}"`, then `[[ -z "$log" ]]` for the usage error and
`[[ ! -f "$log" ]]` for the unreadable file. Count with
`$(wc -l < "$log")`.

### Answer

```bash
log="${1:-}"

if [[ -z "$log" ]]; then
	echo "usage: 1_check_input.sh LOG" >&2
	exit 2
fi

if [[ ! -f "$log" ]]; then
	echo "cannot read $log" >&2
	exit 1
fi

lines=$(wc -l < "$log")
echo "$log: $lines lines"
```

## Task 2: Count each level

Loop over `ERROR WARN INFO`. For each, count the matching lines. The level
column has a space on each side, which a word inside a message usually
doesn't. On `quiet.log`, the ERROR count is zero, so `grep -c` exits 1.

### Closer

`count=$(grep -c " $level " "$1" || true)` inside
`for level in ERROR WARN INFO; do ... done`.

### Answer

```bash
for level in ERROR WARN INFO; do
	count=$(grep -c " $level " "$1" || true)
	echo "$level: $count"
done
```

## Task 3: Rank the failing services

Start the pipeline with `awk`, not `grep`: on a log with no errors, `grep`
exits 1 and `pipefail` fails the whole pipeline. Then count, rank, and
reformat with a second `awk`.

### Closer

`awk '$2 == "ERROR" {print $3}' "$1" | sort | uniq -c | sort -rn | ...`.
After `uniq -c`, the count is `$1` and the service is `$2`:
`awk '{print $2 ": " $1}'`.

### Answer

```bash
awk '$2 == "ERROR" {print $3}' "$1" |
	sort |
	uniq -c |
	sort -rn |
	awk '{print $2 ": " $1}'
```

A `|` at the end of a line lets the pipeline carry on to the next line.

## Task 4: Write the triage report

Define `section` at the top: two `echo`s. Then run your three scripts with
`bash`, each passing `"$1"` along, with a `section` call before the second
and third. The slowest requests are one more pipeline.

### Closer

`section() { echo; echo "== $1 =="; }`. For the last part,
`awk '{print $4 " ms " $3}' "$1" | sort -rn | head -n 3`.

### Answer

```bash
section() {
	echo
	echo "== $1 =="
}

bash 1_check_input.sh "$1"

section "levels"
bash 2_levels.sh "$1"

section "failing services"
bash 3_failing_services.sh "$1"

section "slowest requests"
awk '{print $4 " ms " $3}' "$1" | sort -rn | head -n 3
```
