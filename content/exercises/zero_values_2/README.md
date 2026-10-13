# Zero values: fix the log rotation defaults

Log rotation settings come from a config file where people often leave fields
out. Fix and finish the code that fills in the defaults.

## Learn

### Recap

A variable declared without a value holds its type's zero value: `0` for
`int` and `float64`, `""` for `string`, `false` for `bool`. Missing config
settings arrive as zero values, so you check for zero and swap in a default:

```go
if retries == 0 {
	retries = 3
}
```

A `float64` compares with zero the same way: `if rate == 0 { ... }`.

## Your task

Your team's log shipper reads a rotation config. Some of the default handling
has bugs, and some was never written.

### 1. Fix the log path default

`LogPath` should return `"/var/log/app.log"` when `path` is empty, and `path`
otherwise. Right now the check is backwards.

```go
LogPath("")                  // => "/var/log/app.log"
LogPath("/var/log/api.log")  // => "/var/log/api.log"
```

### 2. Default the sample rate

`SampleRate` returns the fraction of log lines to keep. When `rate` is `0`,
return `0.1`; otherwise return `rate`.

```go
SampleRate(0)    // => 0.1
SampleRate(0.5)  // => 0.5
```

### 3. Finish the rotation summary

`RotationSummary(path, maxSizeMiB, keepFiles)` fills in defaults, then returns
`rotate PATH at NMiB, keep K`:

- an empty `path` becomes `"/var/log/app.log"` (call `LogPath`)
- a `maxSizeMiB` of `0` becomes `100`
- a `keepFiles` of `0` becomes `5`

The starter has a copy-paste bug in the `keepFiles` check, and the
`maxSizeMiB` default is missing.

```go
RotationSummary("", 0, 0)
// => "rotate /var/log/app.log at 100MiB, keep 5"
RotationSummary("/var/log/api.log", 50, 10)
// => "rotate /var/log/api.log at 50MiB, keep 10"
RotationSummary("", 20, 0)
// => "rotate /var/log/app.log at 20MiB, keep 5"
```

## Learn more

- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
- [Go spec: The zero value](https://go.dev/ref/spec#The_zero_value)
