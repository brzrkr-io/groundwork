# Practice: endpoint health checker

Before and after every deploy, someone asks "is everything still up?". In this
practice you build the tool that answers: it checks a list of URLs, retries
flaky ones once, prints a table, and returns an exit code that a CI pipeline
can act on.

## Learn

### Recap

Everything here comes from earlier exercises:

- An `http.Client` with a `Timeout` never hangs forever (*HTTP clients: fetch
  a service status*). `client.Get(url)` is shorthand for building a GET
  request and calling `client.Do`.
- After a successful request, always `defer resp.Body.Close()`. If the request
  fails, there is no body to close.
- `resp.StatusCode` holds the status code. 5xx codes (500 and above) mean the
  server failed, and are often worth one retry.
- Code that reads the clock takes it as a `func() time.Time` field, so tests
  can swap in a fake (*Fakes and dependency injection*). `later.Sub(earlier)`
  is the elapsed `time.Duration`.
- `fmt.Fprintf(&b, ...)` writes formatted text into a `strings.Builder` or any
  other `io.Writer`.
- Exit code 0 means success, anything else means failure
  (*Standard streams: stdin, stdout, stderr and exit codes*).

### How the pieces fit

| Function | Job |
|---|---|
| `NewChecker` | Builds a `Checker` with a real client and the real clock. |
| `Probe` | One GET request: status code or error. |
| `Check` | Probe, retry once on 5xx, measure latency, build a `Result`. |
| `Report` | Turns results into a table. |
| `Run` | Checks every endpoint, prints the report, picks the exit code. |

A real `main` would only glue them together:

```go
c := endpoints.NewChecker(5 * time.Second)
os.Exit(endpoints.Run(c, list, os.Stdout))
```

The tests start local servers with `httptest.NewServer` and give your
`Checker` a fake clock that moves forward a fixed step on every call.

## Your task

Complete the five functions in `exercise.go`. Each builds on the ones before
it.

### 1. Build a checker

`NewChecker(timeout)` returns a `*Checker` whose `Client` is a new
`*http.Client` with `Timeout` set to `timeout`, and whose `Now` is `time.Now`
(the function itself, without calling it).

### 2. Probe one URL

`c.Probe(url)` sends a GET request with `c.Client`.

- If the request fails, return `0` and the error.
- Otherwise close the response body and return `resp.StatusCode` with a `nil`
  error. Any status code counts as a successful probe: a 404 is an answer,
  not an error.

### 3. Check one endpoint

`c.Check(ep)` returns a `Result`:

1. Call `c.Now()` once and remember the start time.
2. Probe `ep.URL`. If the probe succeeded with a status of 500 or higher,
   probe a second time and keep the second answer. Never retry more than once,
   and don't retry when the probe returned an error.
3. Call `c.Now()` a second time. `Latency` is the time since the start.
   Call `c.Now` exactly twice, no matter how many attempts there were.
4. Fill in every field: `Name` and `Expect` from `ep`, `Status` and `Err` from
   the last probe, `Attempts` (1 or 2), and `OK`, which is true only when
   `Err` is `nil` and `Status == ep.Expect`.

```go
// server answers 503, then 200
c.Check(Endpoint{Name: "api", URL: url, Expect: 200})
// => Result{Name: "api", Expect: 200, Status: 200, Attempts: 2, OK: true, ...}
```

### 4. Render the report

`Report(results)` already writes the header. Add one row per result with
`fmt.Fprintf(&b, rowFormat, name, status, latency, result)`, then a summary
line `<ok> of <total> endpoints ok`.

| Column | Value |
|---|---|
| status | the status code, or `-` when `Err` is not nil |
| latency | `r.Latency` (`%s` prints a duration as `120ms` or `2s`) |
| result | `error: <the error text>` when `Err` is not nil, else `ok` when `OK`, else `want <Expect>` |

```text
NAME       STATUS  LATENCY  RESULT
api        200     120ms    ok
payments   503     80ms     want 200
search     -       2s       error: connection refused
1 of 3 endpoints ok
```

Every line, including the last, ends with `\n`. With no results, the report
is the header plus `0 of 0 endpoints ok`. Add the `"strconv"` import if you
use `strconv.Itoa` for the status column.

### 5. Run everything

`Run(c, endpoints, out)` checks each endpoint in order with `c.Check`, writes
`Report(results)` to `out` (for example with `fmt.Fprint`), and returns `0`
if every result is OK, or `1` if any is not. An empty list returns `0`.

## Learn more

- [net/http: Client](https://pkg.go.dev/net/http#Client)
- [net/http/httptest: NewServer](https://pkg.go.dev/net/http/httptest#NewServer)
- [Go by Example: Exit](https://gobyexample.com/exit)
