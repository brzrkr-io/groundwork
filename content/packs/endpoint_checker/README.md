# Practice: endpoint health checker

Before and after every deploy, someone asks "is everything still up?". Here
you build the tool that answers: it checks a list of URLs, retries flaky ones
once, prints a table, and returns an exit code a CI pipeline can act on.

**New here:** nothing; practice.

## Learn

### Recap

- An `http.Client` with a `Timeout` never hangs forever (*HTTP clients: fetch
  a service status*). `client.Get(url)` sends a GET request.
- After a successful request, `defer resp.Body.Close()`. A failed request has
  no body to close.
- `resp.StatusCode` is the status code. 500 and above means the server
  failed, which is often worth one retry.
- Code that reads the clock takes it as a `func() time.Time` field, so tests
  can swap in a fake (*Fakes and dependency injection*).
  `later.Sub(earlier)` is the elapsed `time.Duration`.
- `fmt.Fprintf(&b, ...)` writes formatted text into a `strings.Builder` or any
  `io.Writer`.
- Exit code 0 means success; anything else means failure (*Standard streams:
  stdin, stdout, stderr and exit codes*).

> **Why is it written like this?** `time.Now` (no `()`) is the function
> itself, a value you can store. `time.Now()` *calls* it and gives a time.

### How the pieces fit

| Function | Job |
|---|---|
| `NewChecker` | Builds a `Checker` with a real client and the real clock. |
| `Probe` | One GET request: status code or error. |
| `Check` | Probe, retry once on 5xx, measure latency, build a `Result`. |
| `Report` | Turns results into a table. |
| `Run` | Checks every endpoint, prints the report, picks the exit code. |

A real `main` would just glue them together:

```go
c := endpoints.NewChecker(5 * time.Second)
os.Exit(endpoints.Run(c, list, os.Stdout))
```

The tests start local servers with `httptest.NewServer`, and give your
`Checker` a fake clock that moves forward a fixed step on every call.

## Your task

Complete the five functions in `exercise.go`. Each builds on the ones before.

### 1. Build a checker

**Goal:** `NewChecker(timeout)` returns a `*Checker` with a real client and
the real clock.

- `Client`: a new `*http.Client` whose `Timeout` is the `timeout` parameter.
- `Now`: `time.Now`, the function itself, not called.

**Think it through:**
- Which field gets the parameter, and which gets a function from `time`?

### 2. Probe one URL

**Goal:** `c.Probe(url)` sends one GET with `c.Client` and returns the status
code.

- Request fails: return `0` and the error.
- Otherwise: close the response body and return `resp.StatusCode` with a
  `nil` error. Any status counts as success here: a 404 is an answer, not an
  error.

**Think it through:**
- When is there a body to close, and when isn't there?

### 3. Check one endpoint

**Goal:** `c.Check(ep)` probes `ep.URL` (retrying once on 5xx), times it, and
returns a filled-in `Result`.

1. Call `c.Now()` once: that's the start time.
2. Probe `ep.URL`. If it succeeded with status 500 or higher, probe once
   more and keep the second answer. Never retry more than once, and don't
   retry when the probe returned an error.
3. Call `c.Now()` again. `Latency` is the time since the start. Call
   `c.Now` exactly twice, however many attempts there were.
4. Fill every field: `Name` and `Expect` from `ep`; `Status` and `Err` from
   the last probe; `Attempts` (1 or 2); `OK`, true only when `Err` is `nil`
   and `Status == ep.Expect`.

```go
// server answers 503, then 200
c.Check(Endpoint{Name: "api", URL: url, Expect: 200})
// => Result{Name: "api", Expect: 200, Status: 200, Attempts: 2, OK: true, ...}

// server is down
// => Status: 0, Attempts: 1, Err: (the error), OK: false
```

**Think it through:**
- Which two conditions must both be true before you retry?
- Where do the two `c.Now()` calls go so the retry falls between them?

### 4. Render the report

**Goal:** `Report(results)` adds one table row per result and a summary line
after the header.

Write each row with `fmt.Fprintf(&b, rowFormat, name, status, latency,
result)`. `rowFormat` uses `%s` for every column, so convert the status code
to a string.

| Column | Value |
|---|---|
| name | `r.Name` |
| status | the status code as text, or `-` when `Err` is not nil |
| latency | `r.Latency` (`%s` prints a duration as `120ms` or `2s`) |
| result | `error: <the error text>` when `Err` is not nil; else `ok` when `OK`; else `want <Expect>` |

Then the summary: `<ok> of <total> endpoints ok`.

```text
NAME       STATUS  LATENCY  RESULT
api        200     120ms    ok
payments   503     80ms     want 200
search     -       2s       error: connection refused
1 of 3 endpoints ok
```

Every line, including the last, ends with `\n`. With no results, the report is the
header plus `0 of 0 endpoints ok`. Add the `"strconv"` import if you use
`strconv.Itoa` for the status column.

**Think it through:**
- Which result counts toward the `<ok>` number?
- Mind the space after `error:` and after `want`.

### 5. Run everything

**Goal:** `Run(c, endpoints, out)` checks every endpoint, writes the report
to `out`, and returns the exit code.

Check each endpoint in order with `c.Check`, then write `Report(results)` to
`out` (for example with `fmt.Fprint`). Return `0` if every result is OK, `1`
if any is not. An empty list returns `0`.

**Think it through:**
- What does the exit code start as, and when does it change?

## Learn more

- [net/http: Client](https://pkg.go.dev/net/http#Client)
- [net/http/httptest: NewServer](https://pkg.go.dev/net/http/httptest#NewServer)
- [Go by Example: Exit](https://gobyexample.com/exit)
