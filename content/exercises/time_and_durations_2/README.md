# Time and durations: fix the backup scheduler

A teammate wrote the scheduling helpers for the nightly backup job. Two of
them have classic `time` bugs, and the third is still missing.

## Learn

### Recap

A `time.Time` is an instant and a `time.Duration` is a length of time. You
build durations from unit constants like `30 * time.Second`. Layouts are
written as the reference time `Mon Jan 2 15:04:05 MST 2006`, so `"2006-01-02"`
means "year-month-day".

### Layouts must use the reference time

A layout is not an example date. Writing today's date as a layout gives
garbage, because Go reads each digit as a piece of the reference time:

```go
t := time.Date(2026, 3, 9, 14, 5, 0, 0, time.UTC)
t.Format("2006-01-02") // "2026-03-09"  correct
t.Format("2026-10-11") // "9096-30-33"  wrong: Go read the digits 2, 6, 1, 0... as layout pieces
```

### Durations need a unit

A `time.Duration` counts nanoseconds. Converting a plain number gives you
nanoseconds, not seconds:

```go
attempt := 3
time.Duration(attempt)               // 3ns  (almost certainly a bug)
time.Duration(attempt) * time.Second // 3s
```

You also can't multiply an `int` variable by a `Duration` directly:
`attempt * time.Second` fails with *mismatched types int and time.Duration*.
Convert first, then multiply by the unit. (A constant like `30 * time.Second`
is fine, because untyped constants adapt to the other type.)

### Aligning to a schedule with `Time.Truncate`

`t.Truncate(d)` rounds a **time** down to a multiple of `d`. It counts from
Go's zero time (year 1, midnight UTC), so in UTC, hour and 15-minute
boundaries line up the way you expect:

```go
t := time.Date(2026, 10, 11, 14, 37, 12, 0, time.UTC)
t.Truncate(time.Hour)        // 14:00:00
t.Truncate(15 * time.Minute) // 14:30:00
```

Add one more `d` and you get the next boundary. Everything in this exercise
uses UTC times.

## Your task

The backup job runs on a fixed interval, names its archives by timestamp,
and retries failed uploads.

### 1. Fix the archive name

`ArchiveName(host, at)` should return `<host>-<YYYYMMDD>-<HHMM>.tar.gz`, using
`at` in UTC. It currently uses a real date as its layout. Fix the layout.

```go
ArchiveName("db-01", time.Date(2026, 3, 9, 14, 5, 0, 0, time.UTC))
// => "db-01-20260309-1405.tar.gz"
```

### 2. Fix the retry delay

`RetryDelay(attempt)` should wait `attempt` times 30 seconds, but never more
than 5 minutes. It currently returns nanoseconds. Fix it.

```go
RetryDelay(1)  // => 30s
RetryDelay(4)  // => 2m0s
RetryDelay(10) // => 5m0s  (300s, exactly the cap)
RetryDelay(25) // => 5m0s  (capped)
```

### 3. Find the next run

Write `NextRun(now, every)`: the first boundary of `every` that is strictly
after `now`. Use `now.Truncate(every)` and add `every`. If `now` is exactly on a
boundary, the next run is one full interval later.

```go
now := time.Date(2026, 10, 11, 14, 37, 0, 0, time.UTC)
NextRun(now, time.Hour)        // => 2026-10-11 15:00 UTC
NextRun(now, 15*time.Minute)   // => 2026-10-11 14:45 UTC
NextRun(time.Date(2026, 10, 11, 15, 0, 0, 0, time.UTC), time.Hour)
// => 2026-10-11 16:00 UTC  (already on a boundary)
```

## Learn more

- [time.Time.Truncate](https://pkg.go.dev/time#Time.Truncate)
- [time.Time.Format](https://pkg.go.dev/time#Time.Format)
- [Go blog: Constants](https://go.dev/blog/constants)
