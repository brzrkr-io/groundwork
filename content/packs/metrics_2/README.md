# Prometheus-style metrics: histograms

"Is the API slow?" is the question you'll hear most during an incident.
Counters can't answer it. Here you finish a latency *histogram*, the metric
type Prometheus uses for request durations.

**New here:** histogram buckets, cumulative counts, `time.Time.Sub`,
injecting a clock.

## Learn

### Recap

As in *Prometheus-style metrics: counters and /metrics*: Prometheus scrapes
`GET /metrics`, each family has `# HELP` and `# TYPE` lines then value lines,
and every read or write holds a `sync.Mutex`. This file has a `formatValue`
helper that prints a `float64` the Prometheus way (`0.25`, `3`).

### Why averages lie

99 requests take 0.1s and one takes 30s. The average is about 0.4s and looks
fine, yet one user waited 30 seconds. A histogram answers "how many requests
were slower than 0.5s?"

### Buckets

A histogram has fixed *buckets*, each with an upper bound. Each observation
is counted in one bucket; the histogram also keeps a running sum and count.

```go
h := NewHistogram("http_request_duration_seconds", "Request latency.", []float64{0.1, 0.5})
h.Observe(0.0625) // fits under 0.1
h.Observe(0.25)   // fits under 0.5
h.Observe(1)      // bigger than every bound: the overflow slot
```

### The text format: cumulative

`le` means "less than or equal to". Each bucket line counts **every**
observation at or below its bound, so the numbers only grow. The last
bucket, `le="+Inf"`, counts everything: it equals the total count.

```text
# HELP http_request_duration_seconds Request latency.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1"} 1
http_request_duration_seconds_bucket{le="0.5"} 2
http_request_duration_seconds_bucket{le="+Inf"} 3
http_request_duration_seconds_sum 1.3125
http_request_duration_seconds_count 3
```

From this Prometheus estimates percentiles with `histogram_quantile`, and
the average is `_sum / _count`.

### How Histogram stores its data

`counts` holds plain (**not** cumulative) counts: `counts[i]` is how many
observations landed in bucket `i`. It has one extra slot at the end for
values bigger than every bound:

```text
buckets: [0.1, 0.5]        observations 0.0625, 0.25, 1
counts:  [1,   1,   1]     <- the last slot is the overflow
```

`Write` turns these into cumulative numbers while printing.

> **Why is it written like this?** Storing plain counts is the author's
> choice: `Observe` then touches one slot. `range` over a slice gives you
> both the index and the value (`for i, x := range s`); that's Go syntax.

### Running totals

A running total is a variable declared **before** the loop and added to
inside it:

```go
var total int
for _, page := range pages {
	total += page.Words
	fmt.Println(page.Title, total)
}
```

### Measuring time with an injected clock

Subtracting two times gives a `time.Duration`: `end.Sub(start)`.
`.Seconds()` turns it into a `float64` such as `0.25`.

```go
began := clock()
bake()
fmt.Println(clock().Sub(began).Seconds())
```

`Instrument` takes `now func() time.Time` instead of calling `time.Now`
directly. Production passes `time.Now`; the test passes a fake clock it moves
by hand, so the test is fast and exact.

> **Why is it written like this?** `now` is a parameter holding a function;
> you call it with `now()`. Passing the clock in is a Go testing convention
> called *dependency injection*.

## Your task

The deploy API needs a latency histogram. `Observe` and `Write` are half
written, and nothing measures requests yet. Work in `exercise.go`.

### 1. Put each observation in its bucket

**Goal:** finish `Observe` so it adds 1 to the right slot of `h.counts`.

It already locks the mutex and updates `sum` and `count`. Find the
**first** bucket whose bound is greater than or equal to `v` and add 1 to
that index of `counts`. A value equal to a bound belongs in that bucket. If
`v` is bigger than every bound, add 1 to the last slot,
`counts[len(h.buckets)]`. Each observation goes in exactly one slot.

```go
h := NewHistogram("x", "x", []float64{0.1, 0.5, 1})
h.Observe(0.1)  // counts[0]: 0.1 <= 0.1
h.Observe(0.25) // counts[1]
h.Observe(5)    // counts[3]: the overflow slot
// counts => [1, 1, 0, 1]
```

**Think it through:**
- You loop over `h.buckets`; once you've found the bucket, how do you stop?
- How does the code after the loop know no bucket matched?

### 2. Write cumulative buckets

**Goal:** finish `Write` so it prints one cumulative `_bucket` line per
bound, then the `+Inf` line.

The `HELP`, `TYPE`, `_sum` and `_count` lines are already written. Between
`TYPE` and `_sum`, in bucket order:

```text
<name>_bucket{le="<bound>"} <cumulative count>
<name>_bucket{le="+Inf"} <total count>
```

- `<name>` is `h.name`; `_bucket{le="` and `"} ` are fixed text.
- `<bound>` is printed with `formatValue`; counts with `%d`.
- A bucket's cumulative count is its own count plus all buckets before it.
- The `+Inf` line uses the total count.

See the text format example in Learn. A histogram with no observations
prints `0` everywhere.

**Think it through:**
- Which variable carries the running total between loop iterations?
- Which field already holds the total count for `+Inf`?

### 3. Time every request

**Goal:** finish `Instrument` so it calls `h.Observe` with how long `next`
took, in seconds.

Read the clock with `now()` before `next.ServeHTTP(w, r)` and again after.
Observe the difference in seconds. Leave the `next.ServeHTTP` call as is: the
response goes to the client unchanged.

```text
now() before next: 09:00:00
now() after next:  09:00:00.250
=> h.Observe(0.25)
```

**Think it through:**
- Which time must you save in a variable, and which can you read at the end?
- Which methods turn two times into a `float64` of seconds?

## Learn more

- [Prometheus: Histograms and summaries](https://prometheus.io/docs/practices/histograms/)
- [Prometheus: Metric types](https://prometheus.io/docs/concepts/metric_types/#histogram)
- [time.Time.Sub](https://pkg.go.dev/time#Time.Sub)
