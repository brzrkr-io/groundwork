# Time and durations: fix the backup scheduler

A teammate wrote the scheduling helpers for the nightly backup job. Two of
them have classic `time` bugs, and the third is still missing.

**New here:** `Time.Truncate`; the two classic bugs (a real date as a
layout, a duration without a unit).

## Learn

### Recap

As in *Time and durations: cert expiry, uptime, maintenance windows*: a
`time.Time` is an instant, a `time.Duration` is a length of time built from
units like `30 * time.Second`. Layouts spell out the reference time
`Mon Jan 2 15:04:05 MST 2006`, so `"2006-01-02"` means year-month-day.

### Layouts must use the reference time

A layout is not an example date. Go reads each digit as a piece of the
reference time, so a real date gives garbage:

```go
t := time.Date(2026, 3, 9, 14, 5, 0, 0, time.UTC)
t.Format("2006-01-02") // "2026-03-09"  correct
t.Format("2026-10-11") // "9096-30-33"  wrong: "2" is the day, "1" the month...
t.Format("15h04")      // "14h05"       hour, a literal "h", minute
```

> **Why is it written like this?** The reference time is Go's rule. Letters
> and symbols that aren't layout pieces (like `h` or `-`) are printed as is.

### Durations need a unit

A `time.Duration` counts nanoseconds. Converting a plain number gives
nanoseconds, not seconds:

```go
timeoutSecs := 3
time.Duration(timeoutSecs)               // 3ns  (almost certainly a bug)
time.Duration(timeoutSecs) * time.Second // 3s
```

You can't multiply an `int` variable by a `Duration` directly:
`timeoutSecs * time.Second` fails with *mismatched types int and
time.Duration*. Convert first, then multiply by the unit.

> **Why is it written like this?** Go never mixes types silently: that's a
> rule. A constant like `30 * time.Second` is fine, because untyped
> constants adopt the other type.

### Aligning to a schedule with `Time.Truncate`

`t.Truncate(d)` rounds a **time** down to a multiple of `d`. It counts from
Go's zero time (year 1, midnight UTC), so in UTC, hour and minute boundaries
line up the way you expect:

```go
t := time.Date(2026, 10, 11, 14, 37, 12, 0, time.UTC)
t.Truncate(time.Hour)        // 14:00:00
t.Truncate(10 * time.Minute) // 14:30:00
```

A time already on a boundary doesn't change. Everything here uses UTC.

## Your task

The backup job runs on a fixed interval, names its archives by timestamp,
and retries failed uploads.

### 1. Fix the archive name

**Goal:** fix the layout in `ArchiveName(host, at)` so it returns
`<host>-<YYYYMMDD>-<HHMM>.tar.gz`, using `at` in UTC.

It currently uses a real date as its layout. Only the layout string changes;
`host`, the dashes and `.tar.gz` are already right. Hours use the 24-hour
clock.

```go
ArchiveName("db-01", time.Date(2026, 3, 9, 14, 5, 0, 0, time.UTC))
// => "db-01-20260309-1405.tar.gz"
ArchiveName("web-02", time.Date(2026, 12, 31, 0, 30, 0, 0, time.UTC))
// => "web-02-20261231-0030.tar.gz"
```

**Think it through:**
- Which five pieces does the name need? Find each in the layout table from
  *Time and durations: cert expiry, uptime, maintenance windows*.
- Which pieces sit together with no separator, and where is the `-`?

### 2. Fix the retry delay

**Goal:** fix `RetryDelay(attempt)` so it returns `attempt` × 30 seconds,
never more than `MaxRetryDelay` (5 minutes).

It currently returns nanoseconds. The cap check below it is fine.

```go
RetryDelay(1)  // => 30s
RetryDelay(4)  // => 2m0s
RetryDelay(10) // => 5m0s  (300s, exactly the cap)
RetryDelay(25) // => 5m0s  (capped)
```

**Think it through:**
- Which line computes `delay`? That's the one to change.
- `attempt` is an `int`. What must happen to it before multiplying by a
  unit?
- Which unit turns the 30 into seconds?

### 3. Find the next run

**Goal:** `NextRun(now, every)` returns the first multiple of `every` that is
strictly after `now`.

If `now` is exactly on a boundary, the next run is one full interval later.

```go
now := time.Date(2026, 10, 11, 14, 37, 0, 0, time.UTC)
NextRun(now, time.Hour)      // => 2026-10-11 15:00 UTC
NextRun(now, 15*time.Minute) // => 2026-10-11 14:45 UTC
NextRun(time.Date(2026, 10, 11, 15, 0, 0, 0, time.UTC), time.Hour)
// => 2026-10-11 16:00 UTC  (already on a boundary)
NextRun(time.Date(2026, 10, 11, 23, 50, 0, 0, time.UTC), 30*time.Minute)
// => 2026-10-12 00:00 UTC  (crosses midnight)
```

**Think it through:**
- Round `now` down to its interval first. Which method does that?
- From that boundary, how far is the next one? Which method moves a time
  forward?

## Learn more

- [time.Time.Truncate](https://pkg.go.dev/time#Time.Truncate)
- [time.Time.Format](https://pkg.go.dev/time#Time.Format)
- [Go blog: Constants](https://go.dev/blog/constants)
