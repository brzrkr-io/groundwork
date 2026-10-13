# Prometheus-style metrics: histograms

"Is the API slow?" is the question you'll be asked most during an incident.
Counters can't answer it. In this exercise you finish a latency *histogram*,
the metric type Prometheus uses for request durations.

## Learn

### Recap

As in *Prometheus-style metrics: counters and /metrics*, Prometheus scrapes
`GET /metrics` and reads metric families in a text format: `# HELP` and
`# TYPE` lines, then one `name{label="value"} number` line per value. Metrics
are updated from many goroutines at once, so every read and write holds a
`sync.Mutex`. Numbers are printed with the `formatValue` helper.

### Why averages lie

Say 99 requests take 0.1s and one takes 30s. The average is about 0.4s, which
looks fine, yet one user waited 30 seconds. You want to ask "how many requests
were slower than 0.5s?" A histogram answers that.

### Buckets

A histogram has fixed *buckets*, each with an upper bound. Every observation
is counted in the bucket it falls into, plus a running sum and count:

```go
h := NewHistogram("http_request_duration_seconds", "Request latency.", []float64{0.1, 0.5})
h.Observe(0.0625) // fits under 0.1
h.Observe(0.25)   // fits under 0.5
h.Observe(1)      // bigger than every bound: the overflow slot
```

In the text format, bucket lines are *cumulative*: the `le` label means "less
than or equal to", so each line counts every observation at or below its
bound. The last bucket, `le="+Inf"`, counts everything, so it always equals
the count. Then come `_sum` and `_count`:

```text
# HELP http_request_duration_seconds Request latency.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1"} 1
http_request_duration_seconds_bucket{le="0.5"} 2
http_request_duration_seconds_bucket{le="+Inf"} 3
http_request_duration_seconds_sum 1.3125
http_request_duration_seconds_count 3
```

From these lines Prometheus can estimate percentiles ("95% of requests finish
within 0.4s") with `histogram_quantile`, and the average is `_sum / _count`.

### How Histogram stores its data

`Histogram` keeps one plain (not cumulative) count per bucket in `counts`.
`counts[i]` counts observations that landed in bucket `i`. `counts` has one
extra slot at the end for values bigger than every bound:

```text
Buckets: [0.1, 0.5]        observations 0.0625, 0.25, 1
counts:  [1,   1,   1]     <- the last slot is the overflow
```

`Write` turns those into cumulative numbers when it prints them.

### Measuring time with a fake clock

Durations come from subtracting two times: `end.Sub(start)` is a
`time.Duration`, and `.Seconds()` turns it into a `float64` like `0.25`.
Calling `time.Now` directly would make tests slow and flaky, so
`Instrument` takes a `now func() time.Time`. In production you pass `time.Now`;
the test passes a fake clock that it moves forward by hand.

## Your task

The deploy API needs a latency histogram. `Observe` and `Write` are half
written, and nothing measures requests yet.

### 1. Put each observation in its bucket

Finish `Observe`. It already locks the mutex and updates `sum` and `count`.
Add 1 to the `counts` slot for the **first** bucket whose bound is greater than
or equal to `v`. If `v` is bigger than every bound, add 1 to the last slot,
`counts[len(h.buckets)]`. A value equal to a bound belongs in that bucket.

```go
h := NewHistogram("x", "x", []float64{0.1, 0.5, 1})
h.Observe(0.1)  // counts[0]: 0.1 <= 0.1
h.Observe(0.25) // counts[1]
h.Observe(5)    // counts[3]: the overflow slot
// counts => [1, 1, 0, 1]
```

### 2. Write cumulative buckets

Finish `Write`. The `HELP`, `TYPE`, `_sum` and `_count` lines are written for
you. Between `TYPE` and `_sum`, write one line per bucket, in order, then the
`+Inf` line:

```text
<name>_bucket{le="<bound>"} <cumulative count>
<name>_bucket{le="+Inf"} <total count>
```

The cumulative count of a bucket is its own count plus the counts of every
bucket before it. Print bounds with `formatValue` and counts with `%d`. See
the output in the Learn section for a complete example. A histogram with no
observations prints `0` everywhere.

### 3. Time every request

Finish `Instrument`. It wraps `next`: read the clock with `now()` before
calling `next.ServeHTTP(w, r)` and again after, then call `h.Observe` with the
difference in seconds.

```text
now() before next: 09:00:00
now() after next:  09:00:00.250
=> h.Observe(0.25)
```

The response the handler writes goes to the client unchanged.

## Learn more

- [Prometheus: Histograms and summaries](https://prometheus.io/docs/practices/histograms/)
- [Prometheus: Metric types](https://prometheus.io/docs/concepts/metric_types/#histogram)
- [time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
