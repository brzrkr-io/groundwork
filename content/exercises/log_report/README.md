# Project: access log report

When a service misbehaves, one of the first things you open is its access log.
In this project you build a small analyser for nginx-style access logs: it
parses every line, keeps track of broken lines, works out status codes, busy
paths and latency percentiles, and prints the result as text for humans and
as JSON for dashboards and alerting scripts. It uses everything from this
module: reading lines, regular expressions, time, JSON and error handling.

## Learn

### The log format

Every request is one line:

```text
10.0.0.7 - - [11/Oct/2026:13:55:36 +0000] "GET /api/orders?page=2 HTTP/1.1" 200 512 0.043
```

| Piece | Meaning | `Entry` field |
|---|---|---|
| `10.0.0.7` | client IP address | `IP` |
| `- -` | ident and user name (often `-`, ignored) | |
| `[11/Oct/2026:13:55:36 +0000]` | when the request arrived | `Time` |
| `"GET /api/orders?page=2 HTTP/1.1"` | method, path and protocol | `Method`, `Path` |
| `200` | HTTP status code | `Status` |
| `512` | response size in bytes | `Bytes` |
| `0.043` | how long the request took, in seconds | `Duration` |

This is nginx's standard format with `$request_time` added at the end, a common
tweak in production. `Path` drops the query string: `/api/orders?page=2`
counts as `/api/orders`.

### How the code is organised

The project is one package, `logreport`, split into files by job. Go doesn't
care how you split a package into files. It's only for the humans reading it.

| File | What it holds | Your tasks |
|---|---|---|
| `parse.go` | `Entry`, `ParseLine`, `Log`, `ReadLog` | 1, 2 |
| `stats.go` | `PathCount`, `CountByClass`, `TopPaths`, `Percentile` | 3, 4, 5 |
| `report.go` | `Report`, `Summarize` (done for you), `WriteText`, `WriteJSON` | 6 |
| `logreport_test.go` | one test per task, in task order | read only |
| `testdata/access.log` | a sample log with a blank line and two broken lines | read only |

Data flows in one direction:

```text
io.Reader ──ReadLog──▶ Log ──Summarize──▶ Report ──WriteText / WriteJSON──▶ io.Writer
            (ParseLine          (CountByClass,
             per line)           TopPaths, Percentile)
```

Work through the tasks in order. Every task has its own test, so you see a ✓ as
soon as each piece works, even before the others are done. The scaffolding
(types, signatures, the regular expression, `Summarize`) is already written.

### A few new tools

**nginx timestamps.** As in *Time and durations*, a layout is the reference
time written in your format. The starter defines it for you:
`const timeLayout = "02/Jan/2006:15:04:05 -0700"`.

**Seconds to a Duration without rounding errors.** `0.043` seconds could be
turned into a `time.Duration` with float maths, but floats are not exact, and
you could end up with 42.999999ms. Add an `s` and let `time.ParseDuration` do
it: `time.ParseDuration("0.043s")` is exactly `43ms`.

**Percentiles.** The *p95 latency* is the time that 95% of requests beat. This
project uses the *nearest-rank* method:

1. Sort the durations from fastest to slowest.
2. `rank = ceil(p / 100 × n)`, where `n` is how many durations there are. With
   whole numbers only: `rank := (p*n + 99) / 100`.
3. The answer is the value at position `rank`, counting from 1, so it's
   `sorted[rank-1]`.

For 10 durations, p50 is the 5th fastest and p95 is the 10th, the slowest.

**Sorted map keys.** `maps.Keys(m)` gives you the keys of a map (one by one,
in random order), and `slices.Sorted(...)` collects them into a sorted slice:

```go
for _, class := range slices.Sorted(maps.Keys(byClass)) {
	fmt.Println(class, byClass[class])
}
```

**Writing to an `io.Writer`.** `io.Writer` is the mirror image of `io.Reader`:
anything you can write bytes to, such as a file, a network connection or a
`strings.Builder`. `fmt.Fprintf(w, ...)` works like `Printf` but writes to `w`.
A handy pattern is to build the whole text in a `strings.Builder`, then write it
once and return that single error:

```go
var b strings.Builder
fmt.Fprintf(&b, "requests: %d\n", 16)
_, err := io.WriteString(w, b.String())
return err
```

**Streaming JSON.** `json.NewEncoder(w)` writes JSON straight to a writer.
`enc.SetIndent("", "  ")` works like `MarshalIndent`, and `enc.Encode(v)` adds
a newline after the value.

## Your task

The on-call team wants a quick summary of an access log during incidents. Fill
in the six `TODO`s.

### 1. Parse one line

In `parse.go`, finish `ParseLine(line string) (Entry, error)`. The starter
already matches the line, defines a small `group` helper, and parses the time.
Follow its pattern for the rest.

- `line` is matched with `linePattern.FindStringSubmatch`. It has the named
  groups `ip`, `time`, `method`, `path`, `status`, `bytes` and `duration`;
  `group("status")` looks one up with `linePattern.SubexpIndex`.
- If the line doesn't match, return
  `fmt.Errorf("%w: %q", ErrMalformed, line)`, which prints as
  `malformed line: "the line"`.
- Parse `time` with `timeLayout`. If that fails (for example day `32`), return
  an error. Any message is fine.
- Convert `status` and `bytes` with `strconv.Atoi`, and `duration` with
  `time.ParseDuration(duration + "s")`. Return any error.
- Cut the query string off the path: everything from the first `?` goes.

```go
ParseLine(`10.0.0.7 - - [11/Oct/2026:13:55:36 +0000] "GET /api/orders?page=2 HTTP/1.1" 200 512 0.043`)
// => Entry{IP: "10.0.0.7", Time: 2026-10-11 13:55:36 +0000, Method: "GET",
//          Path: "/api/orders", Status: 200, Bytes: 512, Duration: 43ms}, nil

ParseLine("oops")
// => Entry{}, error "malformed line: \"oops\""
```

### 2. Read a whole log

In `parse.go`, implement `ReadLog(r io.Reader) (Log, error)`.

- Scan `r` line by line, numbering lines from 1. Blank lines (empty after
  `strings.TrimSpace`) still count for numbering, but are skipped without being
  a problem.
- Pass every other line, trimmed, to `ParseLine`. Successful entries go into
  `Entries`, in file order.
- A line that fails does **not** stop the read. Append
  `fmt.Errorf("line %d: %w", n, err)` to `Problems` and carry on. One broken
  line shouldn't hide a thousand good ones.
- If reading itself fails (`sc.Err()`), return `Log{}` and that error.

For `testdata/access.log` you get 16 entries and 2 problems:

```text
line 10: malformed line: "this line was cut in half when logrotate moved the f"
line 19: bad time "32/Oct/2026:13:57:42 +0000": parsing time ...
```

The tests check the first message exactly. For the second, they only check
that it starts with `line 19: `.

### 3. Count status classes

In `stats.go`, implement `CountByClass(entries []Entry) map[string]int`. The
key is the first digit of the status followed by `xx`: 200 and 204 count as
`"2xx"`, 404 as `"4xx"`. With no entries, return an empty map, not `nil`.

```go
CountByClass(entries with statuses 200, 201, 302, 500, 503)
// => map[2xx:2 3xx:1 5xx:2]
```

### 4. Find the busiest paths

In `stats.go`, implement `TopPaths(entries []Entry, n int) []PathCount`.
Count the requests for each `Path`, sort by count from high to low, put paths
with the same count in alphabetical order, and return at most `n` of them.
`slices.SortFunc` with `cmp.Compare` (from *Sorting*) does the sorting.

```go
// paths: /healthz ×3, /api/orders ×3, /login ×2, /metrics ×2
TopPaths(entries, 3)
// => [{/api/orders 3} {/healthz 3} {/login 2}]
```

### 5. Compute a percentile

In `stats.go`, implement `Percentile(durations []time.Duration, p int) time.Duration`
with the nearest-rank method described above. `p` is between 1 and 100.

- Return `0` for an empty slice.
- Don't reorder the caller's slice: sort a copy made with `slices.Clone`.

```go
ds := []time.Duration{90ms, 10ms, 80ms, 20ms, 70ms, 30ms, 60ms, 40ms, 50ms, 100ms}
Percentile(ds, 50) // => 50ms  (rank 5)
Percentile(ds, 90) // => 90ms  (rank 9)
Percentile(ds, 95) // => 100ms (rank 10)
Percentile([]time.Duration{5ms, 1ms, 4ms, 2ms, 3ms}, 50) // => 3ms (rank 3)
```

### 6. Render the reports

In `report.go`, implement both writers. Each returns the error from writing to
`w`.

`WriteText(w io.Writer, r Report) error` writes exactly this layout. Status
classes are sorted by name, top paths keep their order, and durations are
printed with `%v`, so 16ms prints as `16ms` and 2.5 seconds as `2.5s`. Every line
ends with `\n`, including the last one:

```text
requests: 16
skipped: 2
status classes:
  2xx: 11
  3xx: 1
  4xx: 2
  5xx: 2
top paths:
  /api/orders: 6
  /healthz: 4
  /api/orders/42: 2
latency: p50 16ms, p95 2.5s
```

`WriteJSON(w io.Writer, r Report) error` copies `r` into the provided
`jsonReport` struct, with `P50Ms` and `P95Ms` set from `r.P50.Milliseconds()` and
`r.P95.Milliseconds()`, and encodes it with a `json.Encoder` indented by two
spaces. The starter already encodes, but copies only `Requests` and doesn't
indent. The output ends with a newline:

```json
{
  "requests": 16,
  "skipped": 2,
  "by_class": {
    "2xx": 11,
    "3xx": 1,
    "4xx": 2,
    "5xx": 2
  },
  "top_paths": [
    {
      "path": "/api/orders",
      "count": 6
    },
    {
      "path": "/healthz",
      "count": 4
    },
    {
      "path": "/api/orders/42",
      "count": 2
    }
  ],
  "p50_ms": 16,
  "p95_ms": 2500
}
```

These are exactly the numbers for `testdata/access.log`. When all six tests
pass, `ReadLog` followed by `Summarize` turns a real log into this report.

## Learn more

- [nginx: log_format and $request_time](https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format)
- [pkg.go.dev: regexp.Regexp.SubexpIndex](https://pkg.go.dev/regexp#Regexp.SubexpIndex)
- [pkg.go.dev: json.Encoder](https://pkg.go.dev/encoding/json#Encoder)
- [Go by Example: Regular Expressions](https://gobyexample.com/regular-expressions)
