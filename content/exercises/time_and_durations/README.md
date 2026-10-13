# Time and durations: cert expiry, uptime, maintenance windows

Certificates expire, hosts reboot, and maintenance windows open and close.
Ops tools answer "how long until…" and "is it time yet?" all day. Go's
`time` package gives you two types for this: points in time and lengths of
time.

## Learn

### Two types: `time.Time` and `time.Duration`

| Type | What it means | Example |
|---|---|---|
| `time.Time` | A point in time (an *instant*) | 2026-10-11 09:00 UTC |
| `time.Duration` | A length of time | 90 minutes |

A `time.Duration` is really an `int64` counting nanoseconds. You build one
from the unit constants `time.Second`, `time.Minute`, `time.Hour` and so on:

```go
d := 90 * time.Minute
fmt.Println(d)              // 1h30m0s  (Duration has a String method)
fmt.Println(d / time.Hour)  // 1        (integer division drops the rest)
fmt.Println(d % time.Hour)  // 30m0s    (the remainder)
```

Dividing a duration by a duration gives a plain count. Like all integer
division in Go, it *truncates toward zero*: `-36h / 24h` is `-1`, not `-2`.
Convert the result with `int(...)` when you need an `int`.

`time.ParseDuration` reads durations written by humans, such as `"1h30m"`,
`"45s"` or `"2h"`:

```go
d, err := time.ParseDuration("1h30m")
// d == 90 * time.Minute, err == nil
_, err = time.ParseDuration("soon")
// err != nil
```

### Parsing and formatting with layouts

Go doesn't use `%Y-%m-%d` codes. Instead you write a *layout*: the way one
fixed *reference time* would look in your format. The reference time is

```text
Mon Jan 2 15:04:05 MST 2006
```

Memorise it as 1 2 3 4 5 6 7: month **1**, day **2**, hour **3** (or 15),
minute **4**, second **5**, year **6**, zone **-7** (MST is UTC-7).

| Layout piece | Means | Example output |
|---|---|---|
| `2006` | 4-digit year | `2026` |
| `01` | 2-digit month | `10` |
| `02` | 2-digit day | `08` |
| `15` | hour, 24-hour clock | `05` |
| `04` | minute | `30` |
| `05` | second | `00` |
| `MST` | zone name | `UTC` |

`time.RFC3339` is a ready-made layout for the standard timestamp format used
by certificates, APIs and Kubernetes (`2026-10-11T09:00:00Z`):

```go
t, err := time.Parse(time.RFC3339, "2026-10-11T09:00:00Z")
fmt.Println(t.Format("2006-01-02 15:04 MST")) // 2026-10-11 09:00 UTC
```

`t.UTC()` returns the same instant shown in UTC. Use it before formatting so
every host prints the same zone.

### Comparing and doing arithmetic

| Call | Returns | Meaning |
|---|---|---|
| `b.Sub(a)` | `time.Duration` | how long from `a` to `b` (negative if `b` is earlier) |
| `t.Add(d)` | `time.Time` | `t` moved forward by `d` |
| `a.Before(b)` | `bool` | `a` is strictly earlier than `b` |
| `a.After(b)` | `bool` | `a` is strictly later than `b` |
| `a.Equal(b)` | `bool` | same instant (use this, not `==`) |

"At or after `start`" is `!now.Before(start)`.

Durations can also be rounded:

| Call | Result |
|---|---|
| `(100 * time.Minute).Truncate(time.Hour)` | `1h0m0s` (always down) |
| `(100 * time.Minute).Round(time.Hour)` | `2h0m0s` (to nearest) |

The functions in this exercise take `now` as a parameter instead of calling
`time.Now()`. That keeps them testable: the tests pick the exact moment.

## Your task

You're writing helpers for your team's on-call dashboard.

### 1. Days until a certificate expires

`DaysUntilExpiry(now, notAfter)` gets a certificate's expiry time as an
RFC3339 string. Return the number of whole days from `now` until the expiry:
the duration divided by `24 * time.Hour`, truncated toward zero. It is
negative once the certificate has expired.

```go
now := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
DaysUntilExpiry(now, "2026-11-03T08:00:00Z") // => 22, nil  (22 days 23 hours left)
DaysUntilExpiry(now, "2026-10-09T21:00:00Z") // => -1, nil  (expired 36 hours ago)
DaysUntilExpiry(now, "2026-10-11T03:00:00Z") // => 0, nil   (expired 6 hours ago)
```

If `notAfter` can't be parsed, return `0` and the error from `time.Parse`.

### 2. Format an uptime line

`UptimeLine(booted, now)` returns a line in exactly this shape:

```text
up <days>d<hours>h since <booted>
```

- `<days>` is the whole days between `booted` and `now`, `<hours>` is the
  whole hours left over after those days. Minutes and seconds are dropped.
  Both numbers are always shown, even when zero.
- `<booted>` is `booted` converted with `.UTC()` and formatted with the layout
  `"2006-01-02 15:04 MST"`.

```go
booted := time.Date(2026, 10, 8, 5, 30, 0, 0, time.UTC)
now := time.Date(2026, 10, 11, 9, 45, 0, 0, time.UTC)
UptimeLine(booted, now)
// => "up 3d4h since 2026-10-08 05:30 UTC"

UptimeLine(time.Date(2026, 10, 11, 4, 0, 0, 0, time.UTC), now)
// => "up 0d5h since 2026-10-11 04:00 UTC"
```

A `booted` time in another zone is still printed in UTC, for example a boot
at 07:30 in UTC+2 prints as `05:30 UTC`. You'll need `fmt.Sprintf`, so add
`"fmt"` to the imports.

### 3. Check a maintenance window

`InWindow(now, start, length)` reports whether `now` falls inside a
maintenance window. `start` is an RFC3339 string and `length` is a
`time.ParseDuration` string such as `"2h"` or `"30m"`.

The window includes its start and excludes its end:
`start <= now < start + length`.

```go
// Window: 02:00 to 04:00 UTC
now := time.Date(2026, 10, 11, 2, 0, 0, 0, time.UTC)
InWindow(now, "2026-10-11T02:00:00Z", "2h") // => true, nil   (exactly at start)
now = time.Date(2026, 10, 11, 4, 0, 0, 0, time.UTC)
InWindow(now, "2026-10-11T02:00:00Z", "2h") // => false, nil  (exactly at end)
now = time.Date(2026, 10, 11, 1, 59, 0, 0, time.UTC)
InWindow(now, "2026-10-11T02:00:00Z", "2h") // => false, nil  (not open yet)
```

If `start` or `length` can't be parsed, return `false` and that error.

## Learn more

- [Package time](https://pkg.go.dev/time)
- [time.Duration](https://pkg.go.dev/time#Duration)
- [Go by Example: Time Formatting / Parsing](https://gobyexample.com/time-formatting-parsing)
