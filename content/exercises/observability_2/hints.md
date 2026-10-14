## Task 1: Alert on the error ratio

The new rule is a copy of `TargetDown` with a new name, expression, `for`
and severity. Build the expression like `JobsFailing` in Learn, from the
pieces listed in the task.

### Closer

The shape is `sum(rate(<5xx selector>[5m])) / sum(rate(<all requests>[5m])) > 0.05`.
Keep `- alert:` at the same indentation as TargetDown's.

### Answer

```yaml
      - alert: HighErrorRatio
        expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
        for: 10m
        labels:
          severity: page
        annotations:
          summary: More than 5% of deploy API requests are failing
          runbook_url: https://runbooks.example.com/deploy-api/high-error-ratio
```

## Task 2: Alert on slow requests

Same shape as your rule in task 1, with a different expression: a quantile
of the bucket rates, summed so that only `le` is kept.

### Closer

The shape is `histogram_quantile(0.99, sum by (le) (rate(<bucket metric>[5m]))) > 0.5`.
The rule goes under `rules:`, indented like the rules in `1_error_ratio.yml`.

### Answer

```yaml
      - alert: SlowRequests
        expr: histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket[5m]))) > 0.5
        for: 10m
        labels:
          severity: ticket
        annotations:
          summary: Deploy API p99 latency is above 500ms
          runbook_url: https://runbooks.example.com/deploy-api/slow-requests
```

## Task 3: Find the failed traces

One `if` with three conditions, then an `append`. The duplicate check looks
at what is already in `ids`.

### Closer

Join the conditions with `&&`: `line.Msg == "request"`, `line.Status >= 500`
and `!slices.Contains(ids, line.TraceID)`. Add `"slices"` to the imports.

### Answer

```go
if line.Msg == "request" && line.Status >= 500 && !slices.Contains(ids, line.TraceID) {
	ids = append(ids, line.TraceID)
}
```

## Task 4: Find the slowest hop

Two loops over `spans`. The first adds each child's duration to its parent's
total; the second subtracts that total from each span's duration and keeps
the biggest. Skip spans of other traces in both.

### Closer

`childTime := map[string]int{}`, and in the first loop
`childTime[s.ParentID] += s.DurationMS`. In the second,
`own := s.DurationMS - childTime[s.SpanID]`, and replace the best span when
`!ok || own > most`. The named results `hop` and `ok` can hold the best so
far.

### Answer

```go
childTime := map[string]int{}
for _, s := range spans {
	if s.TraceID == traceID && s.ParentID != "" {
		childTime[s.ParentID] += s.DurationMS
	}
}
most := 0
for _, s := range spans {
	if s.TraceID != traceID {
		continue
	}
	if own := s.DurationMS - childTime[s.SpanID]; !ok || own > most {
		hop, most, ok = s, own, true
	}
}
return hop, ok
```
