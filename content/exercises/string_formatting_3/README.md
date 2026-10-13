# Formatting: build an ops report

The morning report shows uptime against the target, open disk alerts, and a
table of the busiest pods. You'll write the three formatters with no
scaffolding beyond the function signatures.

**New here:** width and precision together (`%6.1f`).

## Learn

### Recap

- `fmt.Sprintf` fills verbs in order; values come after the format string.
- `%.2f` rounds a float to 2 decimals and `%%` prints `%`.
- A width pads values: `%-16s` left-aligns a string in 16 characters, and
  `%5d` right-aligns a number in 5.
- `%q` puts a string in double quotes.

### Width and precision together

For floats you can set both: `%6.1f` means "one decimal, padded to at least 6
characters in total". The decimal point counts as a character.

```go
fmt.Sprintf("[%6.1f]", 12.5)  // => "[  12.5]"
fmt.Sprintf("[%6.1f]", 100.0) // => "[ 100.0]"
```

> **Why is it written like this?** **Go requires** the order `%`, flags (like
> `-` or `0`), width, `.precision`, then the verb letter. `%-6.1f` is valid;
> `%.1-6f` is not.

### Reading a long format

Before writing a format, split the expected output into pieces, as in
*Variables: build ops commands*. Mark each piece as fixed text or a value,
then pick a verb for each value. Count the spaces in the example carefully.

## Your task

Write the three report formatters. Add `import "fmt"` on line 4 first.

### 1. Uptime against target

**Goal:** make `UptimeLine` return the service, `: `, the uptime, `% uptime
(target `, the target, then `%)`. Both percentages have 2 decimals.

```go
UptimeLine("checkout", 99.95, 99.9)
// => "checkout: 99.95% uptime (target 99.90%)"

UptimeLine("auth", 100, 99.5)
// => "auth: 100.00% uptime (target 99.50%)"
```

**Think it through:**
- How many `%` signs appear in the output? Each needs `%%`.

### 2. Disk alert

**Goal:** make `DiskAlert` return a logfmt-style alert: host and mount in
double quotes, used with 1 decimal, threshold as a whole number, each number
followed by `%`.

```go
DiskAlert("web-01", "/var", 91.27, 90)
// => `ALERT host="web-01" mount="/var" used=91.3% threshold=90%`

DiskAlert("db-02", "/data", 97, 95)
// => `ALERT host="db-02" mount="/data" used=97.0% threshold=95%`
```

The fixed text: `ALERT host=`, ` mount=`, ` used=`, `% threshold=`, `%`.

**Think it through:**
- `usedPercent` is a `float64` and `thresholdPercent` an `int`. Which verb fits each?

### 3. Pod table row

**Goal:** make `PodRow` return these five pieces in order:

1. the pod name, left-aligned in 16 characters,
2. a space,
3. the CPU percentage with 1 decimal, right-aligned in 5 characters, then `%`,
4. a space,
5. the memory, right-aligned in 5 characters, then ` MiB`.

```go
PodRow("api-7d9f8b-x2k4p", 12.5, 256)
// => "api-7d9f8b-x2k4p  12.5%   256 MiB"

PodRow("worker-5c6d-abcd", 100, 2048)
// => "worker-5c6d-abcd 100.0%  2048 MiB"

PodRow("db-0", 3.04, 16384)
// => "db-0               3.0% 16384 MiB"
```

A 5-character `3.0` is two spaces then `3.0`.

**Think it through:**
- Pieces 1, 3 and 5 each need a width. Which also need a `-` or a precision?

## Learn more

- [Package fmt](https://pkg.go.dev/fmt)
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
- [Google SRE book: Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
