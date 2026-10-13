# Formatting: build an ops report

The morning report shows uptime against the target, open disk alerts, and a
table of the busiest pods. You'll write the three formatters with no
scaffolding beyond the function signatures.

## Learn

### Recap

`fmt.Sprintf` fills verbs in order. `%.2f` rounds a float to 2 decimals and
`%%` prints `%`. A width pads values: `%-16s` left-aligns a string in 16
characters, and `%5d` right-aligns a number in 5.

### Width and precision together

For floats you can set both: `%6.1f` means "one decimal, padded to at least 6
characters in total". The decimal point counts as a character.

```go
fmt.Sprintf("[%6.1f]", 12.5)  // => "[  12.5]"
fmt.Sprintf("[%6.1f]", 100.0) // => "[ 100.0]"
```

## Your task

Write the three report formatters. Add `import "fmt"` first.

### 1. Uptime against target

Write `UptimeLine` so it shows the service, the uptime and the target, both
percentages with 2 decimals:

```go
UptimeLine("checkout", 99.95, 99.9)
// => "checkout: 99.95% uptime (target 99.90%)"

UptimeLine("auth", 100, 99.5)
// => "auth: 100.00% uptime (target 99.50%)"
```

### 2. Disk alert

Write `DiskAlert` so it returns a logfmt-style alert: host and mount in double
quotes, the used percentage with 1 decimal, and the threshold as a whole
number, each followed by `%`:

```go
DiskAlert("web-01", "/var", 91.27, 90)
// => `ALERT host="web-01" mount="/var" used=91.3% threshold=90%`

DiskAlert("db-02", "/data", 97, 95)
// => `ALERT host="db-02" mount="/data" used=97.0% threshold=95%`
```

### 3. Pod table row

Write `PodRow` so it returns:

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

`%5.1f` rounds 3.04 to `3.0`, then pads it to 5 characters: two spaces and
`3.0`.

## Learn more

- [Package fmt](https://pkg.go.dev/fmt)
- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
- [Google SRE book: Service Level Objectives](https://sre.google/sre-book/service-level-objectives/)
