# Zero values: fix the log rotation defaults

Log rotation settings come from a config file where people often leave
fields out. Fix and finish the code that fills in the defaults.

**New here:** nothing; practice with zero values and defaults.

## Learn

### Recap

As in *Zero values: default the config*, an unset value holds its type's
zero value: `0` for `int` and `float64`, `""` for `string`. Check for zero
and swap in a default:

```go
if workers == 0 {
	workers = 4
}
```

A `float64` compares with zero the same way: `if load == 0 { ... }`.

> **Why is it written like this?** `==` means "is equal", `!=` means "is not
> equal" (Go requires these exact symbols). Mixing them up flips the logic,
> and the compiler can't catch it.

### Spotting copy-paste bugs

When two `if` blocks look alike, read each one out loud: "if *this* is
zero, set *this*". The name you check and the name left of `=` should match.

## Your task

Your team's log shipper reads a rotation config. Some default handling has
bugs, and some was never written.

### 1. Fix the log path default

**Goal:** `LogPath(path)` returns `"/var/log/app.log"` when `path` is empty, otherwise `path`.

Right now the check is backwards.

```go
LogPath("")                 // => "/var/log/app.log"
LogPath("/var/log/api.log") // => "/var/log/api.log"
LogPath("/tmp/debug.log")   // => "/tmp/debug.log"
```

**Think it through:**
- Read the `if` line out loud. Does it say "is empty" or "is not empty"?

### 2. Default the sample rate

**Goal:** `SampleRate(rate)` returns `0.1` when `rate` is `0`, otherwise `rate`.

```go
SampleRate(0)   // => 0.1
SampleRate(0.5) // => 0.5
SampleRate(1)   // => 1
```

**Think it through:**
- Same shape as task 1, but the parameter is a `float64`. What is its zero value?

### 3. Finish the rotation summary

**Goal:** `RotationSummary(path, maxSizeMiB, keepFiles)` fills in defaults, then returns `rotate PATH at NMiB, keep K`.

- an empty `path` becomes `"/var/log/app.log"` (already done via `LogPath`)
- a `maxSizeMiB` of `0` becomes `100` (missing)
- a `keepFiles` of `0` becomes `5` (has a copy-paste bug)

The `fmt.Sprintf` line is already correct.

```go
RotationSummary("", 0, 0)                   // => "rotate /var/log/app.log at 100MiB, keep 5"
RotationSummary("/var/log/api.log", 50, 10) // => "rotate /var/log/api.log at 50MiB, keep 10"
RotationSummary("", 20, 0)                  // => "rotate /var/log/app.log at 20MiB, keep 5"
RotationSummary("/var/log/db.log", 0, 3)    // => "rotate /var/log/db.log at 100MiB, keep 3"
```

**Think it through:**
- In the `keepFiles` block, which variable is checked and which is assigned?
- Is there any line that checks `maxSizeMiB`?

## Learn more

- [A Tour of Go: Zero values](https://go.dev/tour/basics/12)
- [Go spec: The zero value](https://go.dev/ref/spec#The_zero_value)
