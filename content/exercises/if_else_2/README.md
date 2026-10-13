# if and else: fix the maintenance script

Your team's maintenance script labels backups, warns about TLS certificates
and picks a log level. A colleague started it, but it doesn't build yet and
one check gives the wrong answers.

This starter does not compile yet; the first job is to read the compiler error.

## Learn

### Recap

As in *if and else*, Go checks `if` / `else if` conditions from top to bottom
and runs only the first block that matches. A `return` inside a branch ends
the function early. A variable made in an `if` short statement
(`if x := ...; x > 5 {`) exists only inside that `if` / `else` chain.

### Reading a compiler error

When code doesn't compile, Go prints the file, line, column and a message:

```text
./exercise.go:12:32: undefined: days
```

Here line 12, column 32 uses a name called `days` that Go can't see at that
point. Gopherlings shows this message in place of the test results.

## Your task

### 1. Fix the compile error in BackupAgeLabel

`BackupAgeLabel` turns a backup's age in hours into a label. It creates `days`
in the `if` short statement and then uses it after the `if` has ended, where
`days` no longer exists. Fix it by declaring `days` on its own line **before**
the `if`, so the whole function can use it. The labels themselves are already
right:

- under 24 hours: `"fresh"`
- otherwise: the number of whole days (`ageHours / 24`), then `"d old"`

```go
BackupAgeLabel(5)  // => "fresh"
BackupAgeLabel(50) // => "2d old"
BackupAgeLabel(24) // => "1d old"
```

### 2. Fix the order of the certificate checks

`CertStatus` should label a TLS certificate by `daysLeft`, the number of days
until it expires:

| `daysLeft` | Return |
|---|---|
| below 0 | `"expired"` |
| 0 to 6 | `"renew now"` |
| 7 to 29 | `"renew soon"` |
| 30 and above | `"valid"` |

The starter checks `daysLeft < 30` first. Any value below 7 is also below 30,
so that first branch wins every time and the other branches can never run.
Reorder the checks so the most urgent one comes first.

```go
CertStatus(-1) // => "expired"
CertStatus(3)  // => "renew now"
CertStatus(12) // => "renew soon"
CertStatus(90) // => "valid"
```

### 3. Finish LogLevelFor

`LogLevelFor` picks a log level for a service:

1. If `debug` is `true`, return `"debug"`, whatever the environment.
2. Otherwise, if `env` is exactly `"production"`, return `"warn"`.
3. Otherwise, return `"info"`.

The first and last rules are already written. Add the middle one.

```go
LogLevelFor("production", true)  // => "debug"
LogLevelFor("production", false) // => "warn"
LogLevelFor("staging", false)    // => "info"
```

## Learn more

- [A Tour of Go: If with a short statement](https://go.dev/tour/flowcontrol/6)
- [Go spec: If statements](https://go.dev/ref/spec#If_statements)
- [Go by Example: If/Else](https://gobyexample.com/if-else)
