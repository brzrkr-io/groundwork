# Named constants

A *magic number* is a bare number in the code whose meaning you have to
guess: is `30` days, seconds, a percentage? Worse, when the same rule
appears in three places as `30`, changing the rule means finding all three
and hoping you didn't miss one or change an unrelated `30`. A named
constant fixes both problems.

**New here:** magic numbers, grouped `const ( ... )` blocks, typed duration
constants.

## Learn

### Recap: constants

As in *Constants*, `const name = value` gives a fixed value a name. The
compiler replaces each use with the value, so a constant costs nothing at
run time.

### Name the meaning, not the number

```go
if attempts > 5 {
```

Five what, and why five? Compare:

```go
const maxRetries = 5

if attempts > maxRetries {
```

Now the line explains itself, and the rule lives in one place. A good
constant name says **what the number means** (`maxRetries`), not what it is
(`five`).

Not every number is magic. `0` in `if count == 0`, or `1` in `i + 1`, means
exactly what it says. Name the numbers that come from a **decision**: a
policy, a limit, a threshold, a port.

### Grouping related constants

A `const ( ... )` block keeps related values together, with one comment
for the group:

```go
// Disk alert levels, in percent used.
const (
	diskWarning  = 80
	diskCritical = 95
)
```

> **Why is it written like this?** **Go requires** a constant's value to be
> known when the code compiles: numbers, strings and expressions of other
> constants, never a function call. Grouping with `( )` and aligning the
> `=` signs is **convention** (gofmt does the aligning). Lower-case names
> keep them private to the package, which is **your choice** here: no
> other package needs them.

### Durations: let the unit speak

`3600 * time.Second` is an hour, but the reader has to do the maths. The
`time` package has the units already:

```go
const scrapeEvery = 15 * time.Second
const retention = 14 * 24 * time.Hour // 14 days
```

`time` has `Nanosecond` up to `Hour`, but no `Day` (days aren't always 24
hours, because of daylight saving), so a day is `24 * time.Hour`.

### How the tests check

`structure_test.go` reads `exercise.go` with Go's `go/ast` package. It
checks that the constants exist and that the functions use no bare numbers
(`0` in `daysLeft < 0` is fine: it means "zero"). The behaviour checks
run first.

## Your task

The certificate checker uses the same thresholds in several places. Name
them, then use the names to roll out a policy change.

| `daysLeft` | `CertStatus` | Due for renewal? | `NextCheck` |
|---|---|---|---|
| below 0 | `"expired"` | yes | 1 hour |
| 0 to 7 | `"critical"` | yes | 1 hour |
| 8 to 30 | `"warning"` | yes | 24 hours |
| more than 30 | `"ok"` | no | 24 hours |

### 1. Name the day thresholds

**Goal:** declare constants `criticalDays` (7) and `warningDays` (30), and use
them instead of every `7` and `30` in `CertStatus`, `RenewalList` and
`NextCheck`.

Declare them at the top of the file, outside any function, where `TODO(1)`
is. Behaviour stays the same:

```go
CertStatus(7)  // => "critical"
CertStatus(8)  // => "warning"
```

**Think it through:**
- How many places use `7`? How many use `30`? Which numbers mean the same thing?
- Which word in the README table goes with 7, and which with 30?

### 2. Name the check intervals

**Goal:** declare constants `urgentInterval` (one hour) and `normalInterval`
(one day) from the `time` package's units, and use them in `NextCheck`.

```go
NextCheck(7) // => 1h0m0s
NextCheck(8) // => 24h0m0s
```

`NextCheck` must have no numbers left in it.

**Think it through:**
- Which `time` constant is exactly one hour?
- `time` has no `Day`. How do you build one?

### 3. Roll out the new warning policy

**Goal:** make certificates warn at 21 days left instead of 30, in both
`CertStatus` and `RenewalList`.

| `daysLeft` | `CertStatus` | Due for renewal? |
|---|---|---|
| 21 | `"warning"` | yes |
| 22 | `"ok"` | no |
| 30 | `"ok"` | no |

**Think it through:**
- After task 1, how many lines do you need to change?
- What would this change have taken before task 1?

## Learn more

- [Go spec: Constant declarations](https://go.dev/ref/spec#Constant_declarations)
- [time package: constants](https://pkg.go.dev/time#pkg-constants)
- [Magic number (programming) on Wikipedia](https://en.wikipedia.org/wiki/Magic_number_%28programming%29)
