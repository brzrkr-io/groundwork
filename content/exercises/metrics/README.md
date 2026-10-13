# Prometheus-style metrics: counters and /metrics

Logs tell you what happened to one request. *Metrics* are numbers that tell
you how the whole service is doing: requests per second, error rate, queue
depth. Dashboards and alerts are built on them. You'll build the core of a
metrics library by hand, so you know what client libraries do for you.

## Learn

### How Prometheus collects metrics

Prometheus is the standard metrics system in the Kubernetes world. It doesn't
wait for your service to send numbers. Every few seconds it *scrapes* your
service: it sends `GET /metrics` and reads the plain-text page that comes back.
Your service's job is to keep its numbers up to date and print them on request.

### Counters and gauges

| Type | Can go down? | Examples |
|---|---|---|
| *Counter* | No, only up (resets to 0 on restart) | requests served, errors, bytes sent |
| *Gauge* | Yes | memory in use, queue length, open connections |

You never read a counter's raw value on a dashboard. Prometheus computes how
fast it grows (`rate(http_requests_total[5m])` = requests per second). That
only works if counters never go down, so adding a negative number to a counter
is a bug, and you ignore it.

### The text format

A scrape returns one *metric family* after another. Each family has a `HELP`
line, a `TYPE` line and one line per value:

```text
# HELP http_requests_total Requests served.
# TYPE http_requests_total counter
http_requests_total{status="200"} 3
http_requests_total{status="500"} 1
```

`status` is a *label*. One family holds a separate counter for every label
value, so you can graph 200s and 500s apart. The line format is
`name{label="value"} number`, followed by a newline.

A word of warning about *cardinality*: every label value is a separate time
series that Prometheus stores. Labels like `status`, `method` or `region` have
a handful of values. Never use a user ID, request ID or full URL as a label:
millions of series will take down your Prometheus.

### Safe for concurrent use

An HTTP server runs each request in its own goroutine, so many goroutines
update the same counter at once. Without a lock, `c.value += v` is a data race
and updates get lost. Protect every read and write with a `sync.Mutex`:

```go
c.mu.Lock()
defer c.mu.Unlock()
c.value += v
```

### Sorted map keys

Map iteration order is random, but a scrape must print lines in a stable
order. `maps.Keys(m)` yields a map's keys, and `slices.Sorted` collects them
into a sorted slice:

```go
for _, key := range slices.Sorted(maps.Keys(m)) {
	fmt.Println(key, m[key])
}
```

## Your task

Your deploy API has no metrics at all. Build a counter, a labeled counter
family and the `/metrics` endpoint. Print counts with
`strconv.FormatFloat(count, 'f', -1, 64)`: it prints `3` for 3 and `2.5` for
2.5. (`'f'` means plain decimal, never `1e+06` style, and `-1` means "as
many digits as needed, and no more".)

### 1. Make Counter safe

Finish `Counter.Add`:

- If `v` is negative, ignore it: the counter doesn't change.
- Otherwise add `v` to the counter, holding `c.mu` while you do.

`Inc` and `Value` are already written. The test calls `Inc` from 50 goroutines
at once and runs with the race detector.

```go
var c Counter
c.Add(2.5)
c.Add(-1)  // ignored
c.Inc()
c.Value()
// => 3.5
```

### 2. Write a counter family

Finish `CounterVec.Write`. The `HELP` and `TYPE` lines are already written.
Add one line per label value, **sorted by label value**, in this format:

```text
<name>{<label>="<label value>"} <count>
```

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

A family with no values yet prints only its `HELP` and `TYPE` lines.

### 3. Serve /metrics

Finish `Handler`. For every request it must:

1. set the `Content-Type` header to `text/plain; version=0.0.4; charset=utf-8`
   (the content type Prometheus expects), and
2. write each family, in the order they were passed in, using `Write`.

The status is 200, which is what you get by not calling `WriteHeader`. You
would mount it with `mux.Handle("GET /metrics", Handler(requests, deploys))`.

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

## Learn more

- [Prometheus: Metric types](https://prometheus.io/docs/concepts/metric_types/)
- [Prometheus: Text exposition format](https://prometheus.io/docs/instrumenting/exposition_formats/)
- [Prometheus: Naming and label best practices](https://prometheus.io/docs/practices/naming/)
- [Go by Example: Mutexes](https://gobyexample.com/mutexes)
