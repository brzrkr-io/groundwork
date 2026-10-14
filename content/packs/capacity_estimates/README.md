# Capacity estimates: QPS, storage and bandwidth

*Practice: containerize and ship a service* closed *Containers and CI/CD*:
you can ship one service. This module adds thinking about whole systems:
load, reliability, caching, sharding and queues.

Before you build or buy anything, you ask: how busy will this be, how much
disk will it fill, and how many servers does it need? You'll turn those
napkin sums into small Go functions you can rerun whenever the numbers change.

**New here:** back-of-the-envelope estimates, peak factor, decimal storage units, bits vs bytes, N+1 spares.

## Learn

### Back-of-the-envelope maths

A *back-of-the-envelope estimate* is a quick sum with round numbers. It won't
be exact, and it doesn't need to be: it tells you whether you need one server
or fifty, gigabytes or petabytes. See the *System design fundamentals* guide
for where estimates fit in a design.

Every estimate here starts from a few inputs: requests per day, how big each
request is, and how long you keep the data.

### From per day to per second

Traffic is usually quoted per day, but servers are sized per second. A day
has 24 × 60 × 60 = **86,400** seconds. *QPS* (queries per second) means
requests per second.

In your head, round 86,400 up to **100,000**. Then "per day" ÷ 100,000 is
roughly QPS, and slightly low, because the real divisor is smaller:

| Per day | Head maths (÷ 100,000) | Exact (÷ 86,400) |
|---|---|---|
| 1 million | 10 QPS | 11.6 QPS |
| 10 million | 100 QPS | 115.7 QPS |
| 86.4 million | 864 QPS | 1,000 QPS |

The code uses the exact 86,400. The head-maths version is for meetings.

### Average vs peak

Traffic is not flat. A CI system is quiet at 3 a.m. and busy at 10 a.m.
The *peak factor* says how much busier the busiest moment is than the
average:

```text
peak QPS = average QPS × peak factor
```

A peak factor of 2 to 5 is common. Size servers and network for the
**peak**, not the average, or the system falls over exactly when people use
it most.

### Storage: decimal units

Storage per day is items per day × bytes per item. Keep it for some days and
it grows by that much every day. This exercise uses **decimal** units, like
disk vendors and cloud bills do:

| Unit | Bytes |
|---|---|
| 1 KB (kilobyte) | 1,000 |
| 1 MB (megabyte) | 1,000,000 (1e6) |
| 1 GB (gigabyte) | 1,000,000,000 (1e9) |
| 1 TB (terabyte) | 1,000,000,000,000 (1e12) |

You'll also meet KiB, MiB and GiB, which use powers of 1,024. For an
estimate the few percent between them don't matter; just pick one and say so.

### Bandwidth: bits, not bytes

Disks are measured in bytes, networks in **bits**. One byte is 8 bits. A
link of 1 Mbit/s (one megabit per second) moves 1,000,000 bits a second.
So bandwidth is bytes per second × 8, then ÷ 1,000,000 for Mbit/s.

### Servers: round up, then add a spare

If one server handles 400 QPS and the peak is 1,000 QPS, you need 2.5
servers, and half a server doesn't exist. Always round **up**. As in
*Rate limiting: fix the bucket*, `math.Ceil` does that and returns a
`float64`, so convert the result to `int`:

```go
math.Ceil(2.5)      // => 3.0
math.Ceil(2.0)      // => 2.0
int(math.Ceil(0.1)) // => 1
```

Then add *spares*. With exactly enough servers, one crash or one rolling
deploy leaves you short at peak. *N+1* means "enough servers (N), plus one
spare".

### Floats in Go

All the inputs are counts (`int`), but the answers have fractions, so these
functions return `float64`. Go never mixes `int` and `float64` in one sum:
convert with `float64(n)` first. Numbers like `1e6` are written in
scientific notation (1 × 10⁶), and `1_000_000` is the same number with
underscores for readability.

```go
pages := 1_200
minutes := 90

pagesPerMinute := float64(pages) / float64(minutes) // => 13.333...
megapages := float64(pages) / 1e6                   // => 0.0012
```

The starter declares constants for the unit sizes (`SecondsPerDay`,
`BytesPerGB`, `BitsPerByte`, `BitsPerMegabit`). They are *untyped*
constants, so they work next to a `float64` without a conversion.

Float answers are rarely exact: after `a, b := 0.1, 0.2`, `a + b` is
`0.30000000000000004`. So the tests accept any answer **within 0.001** of
the expected value.

> **Why is it written like this?** Go requires the `float64(...)`
> conversions: it never converts number types for you. Decimal units and the
> names of the constants are the author's choice. Naming them beats bare
> numbers like `1e9` in the middle of a formula, because the reader sees what
> each number means.

## Your task

Your team is launching a CI log-storage service: every finished CI job
uploads its log, and the service keeps logs for a few weeks. Write the
estimate helpers, then one function that fills in the whole plan.

### 1. Work out average and peak QPS

**Goal:** make `AverageQPS(requestsPerDay)` return requests per second on an average day, and `PeakQPS(requestsPerDay, peakFactor)` return the busiest QPS.

`AverageQPS` spreads `requestsPerDay` evenly over the 86,400 seconds of a day
(use `SecondsPerDay`). `PeakQPS` is the average multiplied by `peakFactor`:
call `AverageQPS` instead of repeating the sum.

```go
AverageQPS(8_640_000)     // => 100
AverageQPS(1_000_000)     // => 11.574... (head maths said "about 10")
AverageQPS(0)             // => 0
PeakQPS(8_640_000, 3)     // => 300
PeakQPS(43_200_000, 2.5)  // => 1250 (average 500)
```

**Think it through:**
- `requestsPerDay` is an `int`, but the answer has fractions. Where does the conversion go?
- Which of `PeakQPS`'s parameters does `AverageQPS` need?

### 2. Estimate storage

**Goal:** make `StorageGB(itemsPerDay, bytesPerItem, days)` return how many decimal gigabytes the data fills after `days` days.

Bytes stored = items per day × bytes per item × days. Divide by
`BytesPerGB` (1e9) to get gigabytes.

```go
StorageGB(200_000, 50_000, 1)   // => 10    (200,000 logs of 50 KB a day)
StorageGB(200_000, 50_000, 365) // => 3650  (a year of them)
StorageGB(1_500, 2_000_000, 30) // => 90    (1,500 files of 2 MB, 30 days)
```

**Think it through:**
- Three `int`s multiplied, and a result with fractions. Convert each one to `float64`, then multiply and divide.

### 3. Estimate bandwidth

**Goal:** make `BandwidthMbps(qps, bytesPerRequest)` return the network traffic in megabits per second.

Bytes per second = `qps` × `bytesPerRequest`. Multiply by `BitsPerByte` (8)
to get bits per second, then divide by `BitsPerMegabit` (1e6).

```go
BandwidthMbps(300, 50_000)    // => 120  (15,000,000 bytes/s = 120,000,000 bits/s)
BandwidthMbps(1250, 2_000)    // => 20
BandwidthMbps(0.5, 1_000_000) // => 4
```

**Think it through:**
- `qps` is already a `float64`. Which parameter still needs converting?
- Bytes to bits: multiply or divide by 8?

### 4. Count servers with a spare

**Goal:** make `ServersNeeded(peakQPS, qpsPerServer, spare)` return how many servers handle `peakQPS`, rounded up, plus `spare` extra servers.

Divide `peakQPS` by what one server handles, round **up** to a whole server
with `math.Ceil`, convert to `int`, then add `spare`. Add `"math"` to the
imports (an `import "math"` line under `package capacity`). With no traffic
you still run the spares.

```go
ServersNeeded(1250, 400, 1) // => 5  (3.125 rounds up to 4, plus 1 spare)
ServersNeeded(1200, 400, 1) // => 4  (exactly 3, plus 1 spare)
ServersNeeded(301, 100, 0)  // => 4
ServersNeeded(0, 400, 1)    // => 1
```

**Think it through:**
- Which step makes 3.125 become 4, and which makes `4.0` an `int`?
- `spare` counts whole servers, not QPS. What goes wrong if you add it to
  `peakQPS` before dividing? Try it with the first example.

### 5. Fill in the plan

**Goal:** make `Estimate(w)` return a `Plan` for the `Workload` `w`, with **one** spare server (N+1).

Fill every field of `Plan` by calling your functions with fields of `w`:

| Plan field | Comes from |
|---|---|
| `PeakQPS` | peak QPS for `w.RequestsPerDay` and `w.PeakFactor` |
| `StorageGB` | `w.RequestsPerDay` logs a day of `w.BytesPerRequest` bytes, kept `w.RetentionDays` days |
| `PeakMbps` | bandwidth at the **peak** QPS, `w.BytesPerRequest` bytes each |
| `Servers` | servers for the **peak** QPS at `w.QPSPerServer` each, plus 1 spare |

```go
Estimate(Workload{
	RequestsPerDay:  4_320_000, // average 50 QPS
	PeakFactor:      3,
	BytesPerRequest: 200_000, // 200 KB logs
	RetentionDays:   30,
	QPSPerServer:    40,
})
// => Plan{PeakQPS: 150, StorageGB: 25920, PeakMbps: 240, Servers: 5}
```

**Think it through:**
- Three fields need the peak QPS. Would a variable for it help?
- You can return a struct literal with every field named: `Plan{PeakQPS: ..., ...}`.

## Learn more

- [Wikipedia: Back-of-the-envelope calculation](https://en.wikipedia.org/wiki/Back-of-the-envelope_calculation)
- [Wikipedia: Data-rate units](https://en.wikipedia.org/wiki/Data-rate_units)
- [Wikipedia: N+1 redundancy](https://en.wikipedia.org/wiki/N%2B1_redundancy)
- [pkg.go.dev: math.Ceil](https://pkg.go.dev/math#Ceil)
