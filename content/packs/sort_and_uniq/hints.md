## Task 1: List the services

Two steps joined with a pipe: first keep only the service column, then sort
the names so each appears once.

### Closer

`cut -d' ' -f3 "$1"` prints the third space-separated column of every line.
Pipe it into `sort` with the flag for unique lines.

### Answer

```bash
cut -d' ' -f3 "$1" | sort -u
```

## Task 2: Find the biggest directories

Each line starts with a number, so you need number order, not text order,
and the biggest first. Then keep only the top of the list.

### Closer

`sort -rn "$1"` sorts by the leading number, biggest first. `head -n` keeps
the first lines of whatever it receives.

### Answer

```bash
sort -rn "$1" | head -n 5
```

## Task 3: Count errors per service

Build the pipe one step at a time and run it after each step: the `ERROR`
lines, then only the service column, then count, then order by count.

### Closer

`grep ERROR "$1" | cut -d' ' -f3` gives one service name per error. `uniq -c`
only counts equal lines that sit next to each other, so `sort` them first,
then `sort -rn` the counted lines.

### Answer

```bash
grep ERROR "$1" | cut -d' ' -f3 | sort | uniq -c | sort -rn
```

## Task 4: Find the top client IPs

It's task 3's counting recipe on a different column, with one more step at
the end to keep N lines. N comes from `$2`, with a default.

### Closer

`limit=${2:-3}` stores `$2`, or 3. The IP is the first space-separated
column, so it's `cut` with field 1. After that comes task 3's
`sort | uniq -c | sort -rn`, and `head -n` with `"$limit"`.

### Answer

```bash
limit=${2:-3}
cut -d' ' -f1 "$1" | sort | uniq -c | sort -rn | head -n "$limit"
```
