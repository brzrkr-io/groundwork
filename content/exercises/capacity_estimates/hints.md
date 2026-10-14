## Task 1: Work out average and peak QPS

Average QPS is requests per day divided by the seconds in a day. Peak QPS is
that average times the peak factor, so `PeakQPS` can call `AverageQPS`.

### Closer

Convert the `int` before dividing: `float64(requestsPerDay) / ...`. The
constant `SecondsPerDay` is untyped, so it works next to a `float64`. In
`PeakQPS`, call `AverageQPS(requestsPerDay)` and multiply.

### Answer

```go
// AverageQPS
return float64(requestsPerDay) / SecondsPerDay

// PeakQPS
return AverageQPS(requestsPerDay) * peakFactor
```

## Task 2: Estimate storage

Multiply the three parameters to get the total bytes, then divide by the
bytes in a gigabyte.

### Closer

Convert each parameter with `float64(...)`, multiply them into a `bytes`
variable, and divide it by `BytesPerGB`.

### Answer

```go
bytes := float64(itemsPerDay) * float64(bytesPerItem) * float64(days)
return bytes / BytesPerGB
```

## Task 3: Estimate bandwidth

Requests per second times bytes per request gives bytes per second. Networks
count bits, so multiply by 8, then divide by a million for megabits.

### Closer

Only `bytesPerRequest` needs `float64(...)`. Then use `BitsPerByte` and
`BitsPerMegabit`: `bytesPerSecond * BitsPerByte / ...`.

### Answer

```go
bytesPerSecond := qps * float64(bytesPerRequest)
return bytesPerSecond * BitsPerByte / BitsPerMegabit
```

## Task 4: Count servers with a spare

Divide the peak by what one server handles. That can be a fraction, so round
it up to a whole server first, and add the spares after.

### Closer

`math.Ceil(peakQPS / qpsPerServer)` rounds up and is still a `float64`, so
wrap it in `int(...)` before adding `spare`. Add `import "math"` under the
`package` line.

### Answer

```go
return int(math.Ceil(peakQPS/qpsPerServer)) + spare
```

Adding `spare` after rounding means a spare is a whole extra server, not a
fraction that rounding swallows.

## Task 5: Fill in the plan

Work out the peak QPS once, then use it for the bandwidth and the servers.
Storage uses requests per day, not the peak.

### Closer

`peak := PeakQPS(w.RequestsPerDay, w.PeakFactor)`, then return a `Plan`
literal whose fields call `StorageGB`, `BandwidthMbps` and `ServersNeeded`
with fields of `w`. The spare is `1`.

### Answer

```go
peak := PeakQPS(w.RequestsPerDay, w.PeakFactor)
return Plan{
	PeakQPS:   peak,
	StorageGB: StorageGB(w.RequestsPerDay, w.BytesPerRequest, w.RetentionDays),
	PeakMbps:  BandwidthMbps(peak, w.BytesPerRequest),
	Servers:   ServersNeeded(peak, w.QPSPerServer, 1),
}
```
