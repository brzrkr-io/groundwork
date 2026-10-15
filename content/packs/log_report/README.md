# Project: access log report

When a service misbehaves, one of the first things you open is its access
log. Here you build a small analyser for nginx-style logs: it parses each
line, notes broken lines, and reports status codes, busy paths and latency
percentiles as text (for humans) and JSON (for dashboards).

**New here:** the nearest-rank percentile, `io.WriteString`, `enc.SetIndent`
and `Duration.Milliseconds`. Everything else combines tools you already know:
reading lines, regular expressions, time, JSON, errors, sorting and golden-file
tests (see "Tools you already have").

## Learn

### The log format

Every request is one line:

```text
10.0.0.7 - - [11/Oct/2026:13:55:36 +0000] "GET /api/orders?page=2 HTTP/1.1" 200 512 0.043
```

| Piece | Meaning | `Entry` field |
|---|---|---|
| `10.0.0.7` | client IP | `IP` |
| `- -` | ident and user (ignored) | |
| `[11/Oct/2026:13:55:36 +0000]` | when the request arrived | `Time` |
| `"GET /api/orders?page=2 HTTP/1.1"` | method, path, protocol | `Method`, `Path` |
| `200` | HTTP status | `Status` |
| `512` | response size in bytes | `Bytes` |
| `0.043` | request time in seconds | `Duration` |

`Path` drops the query string: `/api/orders?page=2` counts as `/api/orders`.

### How the code is organised

One package, `logreport`, split into files by job.

> **Why is it written like this?** Go doesn't care how a package is split
> into files; all files in a folder share one package. The split is the
> author's choice, to help humans find things.

| File | Holds | Your tasks |
|---|---|---|
| `parse.go` | `Entry`, `ParseLine`, `Log`, `ReadLog` | 1, 2 |
| `stats.go` | `PathCount`, `CountByClass`, `TopPaths`, `Percentile` | 3, 4, 5 |
| `report.go` | `Report`, `Summarize` (done), `WriteText`, `WriteJSON` | 6 |
| `logreport_test.go` | one test per task, for tasks 1 to 6 | read only |
| `report_golden_test.go` | your own test of the whole pipeline | 7 |
| `grader_test.go` | how your task 7 test is graded | read only |
| `testdata/access.log` | sample log with a blank line and two broken lines | read only |
| `testdata/report.golden` | the text report `testdata/access.log` should give | read only |

Data flows one way:

```text
io.Reader ──ReadLog──▶ Log ──Summarize──▶ Report ──WriteText / WriteJSON──▶ io.Writer
            (ParseLine          (CountByClass,
             per line)           TopPaths, Percentile)
```

Each task has its own test, so you get a ✓ per piece. Types, signatures, the
regular expression and `Summarize` are already written. In the last task you
write a test yourself, as in the *Testing* module.

### Tools you already have

| You need | Use | Taught in |
|---|---|---|
| read lines from an `io.Reader` | `bufio.Scanner`, `sc.Err()` | *Reading input line by line: a hosts list* |
| text as an `io.Reader` | `strings.NewReader(s)` | *Reading input line by line: a hosts list* |
| pull fields out of a line | named groups in a `regexp` | *Regular expressions: release tags and log lines* |
| parse a timestamp | `time.Parse` with a layout | *Time and durations: cert expiry, uptime, maintenance windows* |
| text like `"90s"` to a `time.Duration` | `time.ParseDuration` | *Time and durations: cert expiry, uptime, maintenance windows* |
| sort by your own rule | `slices.SortFunc`, `cmp.Compare` | *Sorting* |
| a map's keys in order | `slices.Sorted(maps.Keys(m))` | *Sorting* |
| a copy of a slice | `slices.Clone` | *append and copy* |
| print into a writer | `fmt.Fprintf(w, ...)` | *Stringers: a kubectl-style deployment report* |
| JSON straight to a writer | `json.NewEncoder(w)`, `enc.Encode(v)` | *Encoding JSON: fix the alert stream* |
| compare output with a file | a golden file in `testdata/` | *Test fixtures and golden files* |

### nginx timestamps

The starter has the timestamp layout, the reference time written in nginx's
format: `timeLayout = "02/Jan/2006:15:04:05 -0700"`.

### Seconds to a Duration, exactly

Floats aren't exact: `0.043` seconds via float maths might become
`42.999999ms`. Instead, add a unit to the text and let `time.ParseDuration`
read it:

```go
d, err := time.ParseDuration("1.5m") // exactly 90s
```

### Rounding up with whole numbers

Integer division drops the remainder (`7 / 2` is `3`). To round **up**, add
"divisor minus 1" first:

```go
pages := (items + 9) / 10 // 10 per page: 21 items => 3 pages
```

### Percentiles

The *p95 latency* is the time within which 95% of requests finish. This
project uses the *nearest-rank* method:

1. Sort the durations, fastest first.
2. `rank` = `p / 100 × n`, **rounded up** (`n` = how many durations).
3. The answer is the `rank`-th value, counting from 1.

For 10 durations, p50 is the 5th fastest and p95 the 10th (the slowest).

### Writing a string to an io.Writer

`io.Writer` is anything you can write bytes to (a file, a network
connection, a `strings.Builder`). Besides `fmt.Fprintf(w, ...)`, the
function `io.WriteString(w, s)` writes one string and returns the number of
bytes written plus an error.

A handy pattern: build all the text in a `strings.Builder`, write it once,
and return that one error:

```go
var b strings.Builder
fmt.Fprintf(&b, "host: %s\n", "web-1")
_, err := io.WriteString(w, b.String())
return err
```

### Indented JSON from an encoder

`enc.SetIndent("", "  ")` makes a `json.Encoder` indent like
`json.MarshalIndent` from *Encoding JSON: a deployment status report*. Call
it once, before `enc.Encode(v)`.

> **Why is it written like this?** You keep the encoder in a variable
> because you need to call two methods on it (`SetIndent`, then `Encode`).
> The name `enc` is a common convention.

### Durations as whole milliseconds

`d.Milliseconds()` turns a `time.Duration` into an `int64` count of whole
milliseconds: `(2500 * time.Millisecond).Milliseconds()` is `2500`.

## Your task

The on-call team wants a quick log summary during incidents. Fill in the six
`TODO`s.

### 1. Parse one line

**Goal:** in `parse.go`, finish `ParseLine` so it returns a fully filled
`Entry`, or an error.

The starter already matches the line, defines a `group` helper
(`group("status")` gives the `"200"` text), and parses the time.

- **No match:** return an error that wraps `ErrMalformed` and prints as
  `malformed line: ` followed by the line in double quotes.
- **Bad time** (for example day `32`): return an error. Any message works.
- Convert `status` and `bytes` to `int`, and `duration` (seconds) to a
  `time.Duration`. Return any conversion error.
- `Path`: everything from the first `?` is cut off.

```go
ParseLine(`10.0.0.7 - - [11/Oct/2026:13:55:36 +0000] "GET /api/orders?page=2 HTTP/1.1" 200 512 0.043`)
// => Entry{IP: "10.0.0.7", Time: 2026-10-11 13:55:36 +0000, Method: "GET",
//          Path: "/api/orders", Status: 200, Bytes: 512, Duration: 43ms}, nil

ParseLine(`192.168.1.20 - alice [11/Oct/2026:15:55:58 +0200] "POST /login?next=/admin HTTP/2.0" 401 64 1.250`)
// => Path "/login", Status 401, Bytes 64, Duration 1.25s

ParseLine("oops")
// => Entry{}, error "malformed line: \"oops\""
```

**Think it through:**
- The time step (call, check `err`, keep the value) is the pattern. Repeat
  it for each conversion.
- Which `strconv` function turns `"200"` into `200`? What text do you add so
  `ParseDuration` knows the unit?
- Imports needed: `"fmt"`, `"strconv"`, `"strings"`.

### 2. Read a whole log

**Goal:** in `parse.go`, `ReadLog` parses every line into `Entries`, records
bad lines in `Problems`, and keeps going.

- Number lines from 1. Blank lines (empty after `strings.TrimSpace`) count
  for numbering but are skipped, and are not problems.
- Send every other line, trimmed, to `ParseLine`. Good entries go into
  `Entries` in file order.
- A bad line does **not** stop the read. Add an error to `Problems` that
  wraps the `ParseLine` error (with `%w`) and prints as `line N: <error>`.
- If reading itself fails (`sc.Err()`), return `Log{}` and that error.

For `testdata/access.log`: 16 entries (first path `/api/orders`, last
`/metrics`) and 2 problems:

```text
line 10: malformed line: "this line was cut in half when logrotate moved the f"
line 19: bad time "32/Oct/2026:13:57:42 +0000": parsing time ...
```

The first must match exactly; the second only has to start with `line 19: `.

**Think it through:**
- Where does the counter go up so blank lines still count?
- Problems go in a field of the `Log` you return, not in the returned error.
- Import `"bufio"`.

### 3. Count status classes

**Goal:** in `stats.go`, `CountByClass` counts entries by the first digit of
their status, as `"2xx"`, `"3xx"`, ...

200 and 204 both count as `"2xx"`, 404 as `"4xx"`. With no entries, return
an empty map, **not** `nil`.

```go
CountByClass(entries with statuses 200, 201, 204, 302, 404, 500, 503)
// => map[2xx:3 3xx:1 4xx:1 5xx:2]
```

**Think it through:**
- What does integer division by 100 leave of `404`?
- The key has a number part and fixed text `xx`.
- Import `"fmt"`.

### 4. Find the busiest paths

**Goal:** in `stats.go`, `TopPaths(entries, n)` returns at most `n`
`PathCount`s, busiest first.

- Count requests per `Path`.
- Sort by count, high to low. Same count: alphabetical by path.
- Return at most `n` (fewer if there are fewer paths).

```go
// paths: /healthz ×3, /api/orders ×3, /login ×2, /metrics ×2
TopPaths(entries, 1)  // => [{/api/orders 3}]
TopPaths(entries, 3)  // => [{/api/orders 3} {/healthz 3} {/login 2}]
TopPaths(entries, 10) // => all four, same order
```

**Think it through:**
- You can't sort a map. What do you copy the counts into first?
- `slices.SortFunc` with `cmp.Compare`, as in *Sorting*. How do you get
  high-to-low instead of low-to-high?
- Imports: `"cmp"`, `"slices"`.

### 5. Compute a percentile

**Goal:** in `stats.go`, `Percentile(durations, p)` returns the nearest-rank
p-th percentile (`p` is 1 to 100).

- Empty slice: return `0`.
- Don't reorder the caller's slice: sort a copy (`slices.Clone`).

```go
ds := []time.Duration{90ms, 10ms, 80ms, 20ms, 70ms, 30ms, 60ms, 40ms, 50ms, 100ms}
Percentile(ds, 50) // => 50ms  (rank 5)
Percentile(ds, 90) // => 90ms  (rank 9)
Percentile(ds, 95) // => 100ms (rank 10)
Percentile([]time.Duration{5ms, 1ms, 4ms, 2ms, 3ms}, 50) // => 3ms (rank 3)
Percentile([]time.Duration{7ms, 300ms, 12ms}, 100)      // => 300ms
Percentile([]time.Duration{42ms}, 95)                   // => 42ms
```

**Think it through:**
- `p × n / 100`, rounded up: see "Rounding up with whole numbers". What's the
  divisor here?
- Ranks count from 1, slice indexes from 0.

### 6. Render the reports

**Goal:** in `report.go`, `WriteText` and `WriteJSON` write a `Report` to
`w`, and each returns the error from writing.

**`WriteText`** writes exactly this, every line ending in `\n` (the last
too):

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

- Fixed labels; the numbers come from the `Report` fields.
- Indented lines start with 2 spaces.
- Status classes sorted by name; top paths keep their order.
- Durations printed with `%v` (16ms prints as `16ms`, 2.5 seconds as `2.5s`).

**`WriteJSON`** fills the provided `jsonReport` struct with all six fields
from `r`. `P50Ms` and `P95Ms` come from `r.P50.Milliseconds()` and
`r.P95.Milliseconds()`. Encode it indented by two spaces. The starter copies
only `Requests` and doesn't indent. The output ends with a newline:

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

These are the real numbers for `testdata/access.log`.

**Think it through:**
- A map has no order. How do you loop over the classes sorted?
- Writing to a broken `w` must return an error. Where does that error come
  from in each function?
- Imports for `report.go`: `"fmt"`, `"maps"`, `"slices"`, `"strings"`.

### 7. Test the whole pipeline with a golden file

**Goal:** in `report_golden_test.go`, write `TestReportGolden` so it fails
whenever the text report for `testdata/access.log` differs from
`testdata/report.golden`.

The tests for tasks 1 to 6 each check one piece. This test checks that the
pieces fit together, the way the on-call team will run them:

- read `testdata/access.log` with `os.ReadFile`, and pass it to `ReadLog`
  through `strings.NewReader`,
- `Summarize` the log and render it with `WriteText` into a
  `strings.Builder`,
- read `testdata/report.golden` and fail with `t.Errorf` if the text
  differs; stop with `t.Fatal` if reading or parsing fails.

`report.golden` holds the same report as the `WriteText` example in task 6.

`TestGrader` grades your test, as in the *Testing* module, but it changes
the data files instead of the code. Your test must pass on the real files.
It must fail in a scratch folder where `report.golden` has a different
number, where `report.golden` lost its last line, and where `access.log`
lost its last request. Finish tasks 1 to 6 first: your test can only pass
when the whole pipeline works.

**Think it through:**
- As in *Test fixtures and golden files*: get `got`, read `want` from the
  golden file, compare. Which variable holds `got` here?
- `ReadLog` wants an `io.Reader`, and `os.ReadFile` gives you `[]byte`.
  What turns text into a reader?
- Imports: `"os"`, `"path/filepath"`, `"strings"`.

## Learn more

- [nginx: log_format and $request_time](https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format)
- [pkg.go.dev: regexp.Regexp.SubexpIndex](https://pkg.go.dev/regexp#Regexp.SubexpIndex)
- [pkg.go.dev: json.Encoder](https://pkg.go.dev/encoding/json#Encoder)
- [Go by Example: Regular Expressions](https://gobyexample.com/regular-expressions)
