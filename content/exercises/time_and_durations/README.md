# Time and durations: cert expiry, uptime, maintenance windows

Certificates expire, hosts reboot, and maintenance windows open and close.
Ops tools answer "how long until…" and "is it time yet?" all day. Go's
`time` package gives you two types for this: points in time and lengths of
time.

**New here:** `time.Time`, `time.Parse`, layouts, `time.ParseDuration`,
`Sub`, `Add`, `Before`, `After`, `time.Duration` arithmetic.

## Learn

### Two types: `time.Time` and `time.Duration`

| Type | What it means | Example |
|---|---|---|
| `time.Time` | A point in time (an *instant*) | 2026-10-11 09:00 UTC |
| `time.Duration` | A length of time | 90 minutes |

You met `time.Duration` in *Fakes: a fake clock for retries*. Underneath,
it's an `int64` counting nanoseconds. Build one from unit constants such as
`time.Second`, `time.Minute`, `time.Hour`:

```go
d := 150 * time.Minute
fmt.Println(d)             // 2h30m0s  (Duration has a String method)
fmt.Println(d / time.Hour) // 2        (integer division drops the rest)
fmt.Println(d % time.Hour) // 30m0s    (the remainder)
```

Duration ÷ duration gives a plain count. Like all integer division in Go, it
*truncates toward zero*: `-36h / 24h` is `-1`, not `-2`. Wrap it in
`int(...)` when you need an `int`.

> **Why is it written like this?** `150 * time.Minute` works because `150`
> is an untyped constant and adopts the `Duration` type (a Go rule). A name
> like `const week = 7 * 24 * time.Hour` is the author's choice; Go has no
> `time.Day`.

`time.ParseDuration` reads durations written by humans:

```go
d, err := time.ParseDuration("1h30m") // d == 90 * time.Minute, err == nil
_, err = time.ParseDuration("soon")   // err != nil
```

### Parsing and formatting with layouts

Go doesn't use `%Y-%m-%d` codes. You write a *layout*: how one fixed
*reference time* would look in your format. The reference time is

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

Everything else in a layout (`-`, `:`, spaces) is printed as is.

`time.RFC3339` is a ready-made layout for the standard timestamp used by
certificates, APIs and Kubernetes (`2026-10-11T09:00:00Z`):

```go
t, err := time.Parse(time.RFC3339, "2026-10-11T09:00:00Z")
fmt.Println(t.Format("02/01/2006 at 15:04")) // 11/10/2026 at 09:00
```

`t.UTC()` returns the same instant shown in UTC. Use it before formatting so
every host prints the same zone.

> **Why is it written like this?** The reference-time layout is Go's rule:
> other dates won't work. `time.Parse` takes the layout first, the text
> second; that order is fixed.

### Comparing and doing arithmetic

| Call | Returns | Meaning |
|---|---|---|
| `b.Sub(a)` | `time.Duration` | how long from `a` to `b` (negative if `b` is earlier) |
| `t.Add(d)` | `time.Time` | `t` moved forward by `d` |
| `a.Before(b)` | `bool` | `a` is strictly earlier than `b` |
| `a.After(b)` | `bool` | `a` is strictly later than `b` |
| `a.Equal(b)` | `bool` | same instant (use this, not `==`) |

"Strictly" means equal times give `false`. So "`deadline` has not passed
yet, or is right now" is `!now.After(deadline)`.

The functions here take `now` as a parameter instead of calling `time.Now()`.
That keeps them testable: the tests pick the exact moment.

## Your task

You're writing helpers for your team's on-call dashboard.

### 1. Days until a certificate expires

**Goal:** `DaysUntilExpiry(now, notAfter)` returns the whole days from `now`
until the RFC3339 time `notAfter`.

- Whole days = the duration ÷ `24 * time.Hour`, truncated toward zero.
- Negative once the certificate has expired.
- If `notAfter` can't be parsed, return `0` and the error from `time.Parse`.

```go
now := time.Date(2026, 10, 11, 9, 0, 0, 0, time.UTC)
DaysUntilExpiry(now, "2026-11-03T08:00:00Z") // => 22, nil  (22 days 23 hours left)
DaysUntilExpiry(now, "2026-11-10T09:00:00Z") // => 30, nil  (exactly 30 days)
DaysUntilExpiry(now, "2026-10-11T03:00:00Z") // => 0, nil   (expired 6 hours ago)
DaysUntilExpiry(now, "2026-10-09T21:00:00Z") // => -1, nil  (expired 36 hours ago)
DaysUntilExpiry(now, "2026-11-03")           // => 0, error (no time part)
```

**Think it through:**
- Which parameter is the text to parse, and which layout fits it?
- Duration from `now` until the expiry: which time calls `Sub` on which?

### 2. Format an uptime line

**Goal:** `UptimeLine(booted, now)` returns
`up <days>d<hours>h since <booted>`.

- `<days>`: whole days between `booted` and `now`.
- `<hours>`: whole hours left over after those days. Minutes and seconds are
  dropped. Both numbers always appear, even when `0`.
- `<booted>`: `booted` converted with `.UTC()`, formatted with the layout
  `"2006-01-02 15:04 MST"`.
- Fixed text: `up `, `d`, `h since `.

```go
booted := time.Date(2026, 10, 8, 5, 30, 0, 0, time.UTC)
now := time.Date(2026, 10, 11, 9, 45, 0, 0, time.UTC)
UptimeLine(booted, now)
// => "up 3d4h since 2026-10-08 05:30 UTC"

UptimeLine(time.Date(2026, 10, 11, 4, 0, 0, 0, time.UTC), now)
// => "up 0d5h since 2026-10-11 04:00 UTC"
```

A boot at 07:30 in UTC+2 prints as `05:30 UTC`. Add `"fmt"` to the imports
for `fmt.Sprintf`.

**Think it through:**
- Get the uptime as one duration first. Then: whole days with `/`, the rest
  with `%`.
- How does the `d % time.Hour` example in Learn get "what's left over"?
- Mind the spaces in the fixed text: there's one after `up` and around
  `since`.

### 3. Check a maintenance window

**Goal:** `InWindow(now, start, length)` reports whether `now` is inside the
window `start <= now < start + length`.

- `start` is an RFC3339 string; `length` is a `time.ParseDuration` string
  such as `"2h"` or `"30m"`.
- The start is inside the window; the end is not.
- If `start` or `length` can't be parsed, return `false` and that error.

```go
// Window: 02:00 to 04:00 UTC
InWindow(time.Date(2026, 10, 11, 2, 0, 0, 0, time.UTC), "2026-10-11T02:00:00Z", "2h")
// => true, nil   (exactly at start)
InWindow(time.Date(2026, 10, 11, 4, 0, 0, 0, time.UTC), "2026-10-11T02:00:00Z", "2h")
// => false, nil  (exactly at end)
InWindow(time.Date(2026, 10, 11, 1, 59, 0, 0, time.UTC), "2026-10-11T02:00:00Z", "2h")
// => false, nil  (not open yet)
```

**Think it through:**
- Two strings to parse, two errors to check. Which function reads each?
- How do you get the end time from the start and the length?
- `Before` is strict. Which edge needs "not before", and which needs
  "before"?

## Learn more

- [Package time](https://pkg.go.dev/time)
- [time.Duration](https://pkg.go.dev/time#Duration)
- [Go by Example: Time Formatting / Parsing](https://gobyexample.com/time-formatting-parsing)
