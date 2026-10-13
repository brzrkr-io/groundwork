# Prometheus-style metrics: counters and /metrics

Logs tell you what happened to one request. *Metrics* are numbers about the
whole service: requests per second, error rate, queue depth. Dashboards and
alerts are built on them. You'll build the core of a metrics library by hand.

**New here:** counters, the Prometheus text format, `sync.Mutex` in practice,
sorted map keys, `strconv.FormatFloat`, setting a response header.

## Learn

### How Prometheus collects metrics

Prometheus doesn't wait for your service to send numbers. Every few seconds
it *scrapes* you: it sends `GET /metrics` and reads the plain-text page that
comes back. Your job is to keep the numbers current and print them on request.

### Counters and gauges

| Type | Can go down? | Examples |
|---|---|---|
| *Counter* | No, only up (resets to 0 on restart) | requests served, errors |
| *Gauge* | Yes | memory in use, queue length |

Dashboards show how fast a counter grows (`rate(...)` = per second). That only
works if counters never go down, so a negative add is a bug: ignore it.

### The text format

A scrape returns *metric families*. Each has a `HELP` line, a `TYPE` line,
and one line per value:

```text
# HELP jobs_done_total Jobs finished.
# TYPE jobs_done_total counter
jobs_done_total{queue="email"} 7
jobs_done_total{queue="resize"} 2
```

`queue` is a *label*: one family holds a separate counter per label value.
Each value line is `name{label="value"} number` plus a newline.

Keep labels to a handful of values (`status`, `method`, `region`). A user ID
or full URL as a label creates millions of series and takes Prometheus down.

### Printing quotes and numbers

To print a `"` inside a Go string, escape it as `\"`. `fmt.Fprintf(w, ...)`
works like `Printf` but writes to `w`:

```go
fmt.Fprintf(w, "%s=\"%s\"\n", "env", "prod")
// env="prod"
```

For the count, use `strconv.FormatFloat(x, 'f', -1, 64)`. It prints `3` for
3 and `2.5` for 2.5. `'f'` means plain decimal (never `1e+06`), `-1` means
"only as many digits as needed", `64` means `x` is a `float64`.

> **Why is it written like this?** The `\"` escape is required by Go: a bare
> `"` would end the string. `\n` is Go's newline escape.

### Safe for concurrent use

An HTTP server runs each request in its own goroutine, so many goroutines
update the same counter at once. Without a lock, `+=` is a *data race* and
updates get lost. A `sync.Mutex` lets one goroutine in at a time:

```go
func (b *Bank) Deposit(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.balance += n
}
```

> **Why is it written like this?** `defer` runs `Unlock` when the function
> returns, however it returns. That's the Go convention: lock, then
> immediately `defer` the unlock, so you can't forget it.

### Sorted map keys

Map order is random, but a scrape must print lines in a stable order.
`maps.Keys(m)` yields the keys; `slices.Sorted` turns them into a sorted
slice you can range over:

```go
for _, name := range slices.Sorted(maps.Keys(ages)) {
	fmt.Println(name, ages[name])
}
```

### Response headers

`w.Header().Set("Name", "value")` sets a response header. Set headers before
writing the body: once the body starts, headers are already sent.

## Your task

Your deploy API has no metrics. Build a counter, a labeled counter family,
and the `/metrics` endpoint in `exercise.go`.

### 1. Make Counter safe

**Goal:** finish `Counter.Add` so it ignores negative `v` and adds anything
else while holding `c.mu`.

- `v` negative → return without changing anything.
- Otherwise lock `c.mu`, add `v` to `c.value`, unlock.

`Inc` and `Value` are already written. The test calls `Inc` from 50
goroutines at once (5000 total) and runs with the race detector.

```go
var c Counter
c.Add(2.5)
c.Add(-1)  // ignored
c.Inc()
c.Value()
// => 3.5
```

**Think it through:**
- Look at `Value` in the same file: it already shows the locking pattern.
- Where does the negative check go: before or after locking?

### 2. Write a counter family

**Goal:** finish `CounterVec.Write` so it writes one line per label value,
sorted by label value.

The `HELP` and `TYPE` lines are already written. Each value line is:

```text
<name>{<label>="<label value>"} <count>
```

`<name>` and `<label>` come from the `CounterVec` fields `name` and `label`.
The label value is a key of `v.counters`; the count is that counter's
`Value()`, printed with `strconv.FormatFloat`. Braces, `=`, the quotes and
the space are fixed text. Add imports `maps`, `slices` and `strconv`.

```go
v := NewCounterVec("http_requests_total", "Requests served.", "status")
v.Inc("500")
v.Inc("200")
v.Inc("200")
v.Write(os.Stdout)
```

```text
# HELP http_requests_total Requests served.
# TYPE http_requests_total counter
http_requests_total{status="200"} 2
http_requests_total{status="500"} 1
```

A family with no values prints only its `HELP` and `TYPE` lines.

**Think it through:**
- How many `%s` does your format string need, and in which order?
- Which characters in the line need a `\` escape?

### 3. Serve /metrics

**Goal:** finish `Handler` so every request gets the Prometheus content type
and every family's output.

1. Set the `Content-Type` header to `text/plain; version=0.0.4; charset=utf-8`.
2. Call `Write` on each family, in the order they were passed in.

Status is 200, which you get by not calling `WriteHeader`. You'd mount it as
`mux.Handle("GET /metrics", Handler(requests, deploys))`.

```text
GET /metrics
200, Content-Type: text/plain; version=0.0.4; charset=utf-8

# HELP http_requests_total Requests served.
# TYPE http_requests_total counter
http_requests_total{status="200"} 2
# HELP deploys_total Deploys started.
# TYPE deploys_total counter
deploys_total{region="eu-west-1"} 1
```

**Think it through:**
- `families` is a variadic parameter: inside the function it's a slice.
- What can each family write to? Look at `Write`'s parameter type and at `w`.

## Learn more

- [Prometheus: Metric types](https://prometheus.io/docs/concepts/metric_types/)
- [Prometheus: Text exposition format](https://prometheus.io/docs/instrumenting/exposition_formats/)
- [Prometheus: Naming and label best practices](https://prometheus.io/docs/practices/naming/)
- [Go by Example: Mutexes](https://gobyexample.com/mutexes)
