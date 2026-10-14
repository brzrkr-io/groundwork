# Observability: alert rules and an incident walk-through

Metrics are what wake you up; traces and logs are how you find the cause.
Here you write the two alerts every HTTP service needs, then follow one
incident from the alert to the log lines and the slow hop.

**New here:** PromQL basics (`rate`, `sum`, `sum by`, `histogram_quantile`,
`=~` label matching), Prometheus alert rules (`for`, `labels`,
`annotations`), a span's own time.

## Learn

### Recap

As in *Observability: trace IDs through logs and calls*: a *trace* is one
request's journey, each hop is a *span*, every span has the same trace ID,
and its parent ID names the span that called it. Every access log line
carries `trace_id`. As in *Decoding JSON: fix the kubectl pod reader*, a
`json.Decoder` reads one JSON value after another until `io.EOF`.
`slices.Contains(s, v)` reports whether `s` already holds `v`.

### rate, by hand

A counter only goes up, so its raw value says little. Two scrapes five
minutes (300 s) apart:

```text
09:05  http_requests_total{status="200"} 12000   {status="503"} 40
09:10  http_requests_total{status="200"} 14700   {status="503"} 340
```

`rate(...[5m])` is the increase per second: 2700/300 = 9 for `200`, 300/300
= 1 for `503`. The error ratio is 1 / (9 + 1) = 10%.

### PromQL in five pieces

Prometheus answers questions in PromQL, its query language. With a worker's
metrics:

| Piece | Example | Means |
|---|---|---|
| selector | `jobs_total{result=~"fail.*"}` | series whose `result` label matches the regular expression |
| `rate(...[10m])` | `rate(jobs_total[10m])` | per-second increase of a counter over the last 10 minutes |
| `sum` | `sum(rate(jobs_total[10m]))` | add every series into one number |
| `sum by (le)` | `sum by (le) (rate(job_seconds_bucket[10m]))` | add up, but keep one result per `le` bucket |
| comparison | `... > 0.1` | keep only results above 0.1; no result, no alert |

A ratio divides two sums: "share of jobs that failed".

```promql
sum(rate(jobs_total{result=~"fail.*"}[10m])) / sum(rate(jobs_total[10m])) > 0.1
```

Always sum both sides. Without `sum`, Prometheus divides series with equal
labels one by one, so every failing series is divided by itself: 1.

For a histogram from *Prometheus-style metrics: histograms*,
`histogram_quantile(0.95, <bucket rates summed by le>)` estimates the
95th-percentile duration: 95% of jobs were faster.

### Alert rules

Prometheus reads rules from YAML files and checks each one every evaluation
interval (often a minute):

```yaml
groups:
  - name: worker
    rules:
      - alert: JobsFailing
        expr: sum(rate(jobs_total{result=~"fail.*"}[10m])) / sum(rate(jobs_total[10m])) > 0.1
        for: 15m
        labels:
          severity: ticket
        annotations:
          summary: More than 10% of jobs are failing
          runbook_url: https://runbooks.example.com/worker/jobs-failing
```

| Field | Meaning |
|---|---|
| `alert` | the alert's name, CamelCase |
| `expr` | PromQL; the alert is active while it returns a result |
| `for` | how long it must stay active before it fires; skips short blips |
| `labels.severity` | routing: `page` wakes someone, `ticket` waits for work hours |
| `annotations` | for humans: a `summary`, and a `runbook_url` with the fix-it steps |

As in the *Data formats and parsing* guide: indentation is structure (spaces,
never tabs) and `- ` starts a list item. The `expr` needs no quotes. In a
real repository, `promtool check rules 1_error_ratio.yml` checks a rule
file; here the test parses it.

### A span's own time

A span's duration **includes** the spans it called. The first span always
looks slowest, because it contains everything:

```text
deploy-api   POST /deploys   4300ms   own: 4300 - 4100 = 200
└ deployer   POST /deploy    4100ms   own: 4100 - 3800 = 300
  └ registry push manifest   3800ms   own: 3800              <- the slow hop
```

A span's *own time* is its duration minus the durations of its **direct**
children (spans whose `ParentID` is its `SpanID`). The hop with the most own
time is where to look.

### The incident

At 09:12 `HighErrorRatio` fires. Then:

1. **Logs:** `testdata/app.log` holds the JSON access logs of `deploy-api`
   and `deployer`. Open it: which `request` lines have status 500 or more?
   The same trace shows up twice, once per service.
2. **Traces:** `testdata/spans.json` holds the spans the tracing backend
   exported. For each failed trace, which hop has the most own time?

`Investigate`, already written, does both steps with your functions and
prints one line per failed trace.

## Your task

Write the deploy API's two alerts, then the two functions `Investigate`
needs. Tasks 1 and 2 are YAML files; tasks 3 and 4 are in `exercise.go`.

### 1. Alert on the error ratio

**Goal:** add the rule `HighErrorRatio` to `1_error_ratio.yml`, below the
finished `TargetDown` example.

The metric is `http_requests_total` from *Prometheus-style metrics: counters
and /metrics*. The `expr`, "more than 5% of requests failed over the last 5
minutes", is built from:

- top: the `sum` of `rate(http_requests_total{status=~"5.."}[5m])` (5xx only)
- `/`
- bottom: the `sum` of `rate(http_requests_total[5m])` (all requests)
- then `> 0.05`

| Field | Value |
|---|---|
| `for` | `10m` |
| `labels.severity` | `page` |
| `annotations.summary` | any one-line text |
| `annotations.runbook_url` | a link starting with `https://` |

The test parses the YAML, checks every field, and checks that `expr`
contains those pieces, has every `rate` inside a `sum`, and ends with
`> 0.05`. Spaces don't matter.

**Think it through:**
- Copy `TargetDown`'s shape and indentation. Which lines change?
- Compare with `JobsFailing` in Learn: what plays the role of
  `result=~"fail.*"` here?

### 2. Alert on slow requests

**Goal:** add the rule `SlowRequests` to `2_slow_requests.yml`.

The histogram is `http_request_duration_seconds` from *Prometheus-style
metrics: histograms*. The `expr`, "the 99th-percentile latency over the last
5 minutes is above 0.5 seconds", is built from:

- `histogram_quantile(0.99, ...)` of
- the `sum by (le)` of `rate(http_request_duration_seconds_bucket[5m])`
- then `> 0.5`

| Field | Value |
|---|---|
| `for` | `10m` |
| `labels.severity` | `ticket` |
| `annotations.summary` | any one-line text |
| `annotations.runbook_url` | a link starting with `https://` |

The same checks as task 1 apply.

**Think it through:**
- The file has the group already. How deep is a rule indented under
  `rules:` in `1_error_ratio.yml`?
- Which metric name ends in `_bucket`, and why does `le` have to survive the
  `sum`?

### 3. Find the failed traces

**Goal:** finish `FailedTraceIDs` so it returns the trace ID of every failed
request, each once, in the order they first appear.

The loop that decodes one `logLine` after another is already written. At the
`TODO(3)`, append `line.TraceID` to `ids` when all of these hold:

- `line.Msg` is `request`
- `line.Status` is 500 or more
- `ids` doesn't hold that trace ID yet (add the `"slices"` import)

```text
{"msg":"request","status":502,"trace_id":"t-b"}
{"msg":"registry push failed","status":500,"trace_id":"t-x"}   <- not a request line
{"msg":"request","status":404,"trace_id":"t-c"}                <- not 500 or more
{"msg":"request","status":503,"trace_id":"t-a"}
{"msg":"request","status":500,"trace_id":"t-b"}                <- t-b already found
// => ["t-b", "t-a"], nil
```

On `testdata/app.log` the answer is
`["4bf92f3577b34da6a3ce929d0e0e4736", "a3ce929d0e0e47364bf92f3577b34da6"]`.

**Think it through:**
- Three conditions, all needed: which operator joins them?
- How do you write "doesn't hold" with `slices.Contains`?

### 4. Find the slowest hop

**Goal:** finish `SlowestHop(spans, traceID)` so it returns the span of
`traceID` with the most own time, and `true`.

- Only spans whose `TraceID` is `traceID` count; ignore the rest.
- Own time = `DurationMS` minus the `DurationMS` of every span whose
  `ParentID` is this span's `SpanID`.
- A tie: the span that comes first in `spans` wins.
- No span has `traceID`: return `Span{}, false`.

```go
// a: deploy-api 1000ms; b: deployer 900ms, parent a;
// c: registry 200ms, parent b; d: kubernetes 100ms, parent b
SlowestHop(spans, "t1") // => span b (900 - 200 - 100 = 600 own), true
SlowestHop(spans, "t9") // => Span{}, false
```

Then `Investigate("testdata/app.log", "testdata/spans.json")` gives:

```text
4bf92f3577b34da6a3ce929d0e0e4736: registry push manifest
a3ce929d0e0e47364bf92f3577b34da6: deployer render manifests
```

**Think it through:**
- First pass: a map from span ID to the total time of its children. Which
  field of a child is the key?
- Second pass: keep the best span so far. How do you make the first span
  count even when its own time is 0?

## Learn more

- [Prometheus: Alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)
- [Prometheus: Querying basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Prometheus: histogram_quantile](https://prometheus.io/docs/prometheus/latest/querying/functions/#histogram_quantile)
- [OpenTelemetry: Traces](https://opentelemetry.io/docs/concepts/signals/traces/)
