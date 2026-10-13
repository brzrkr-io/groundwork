# Formatting: fix the bot messages

The scaling bot's messages are broken: one won't even build, one is missing,
and incident tickets need properly padded IDs.

**New here:** vet errors, zero padding (`%05d`).

## Learn

### Recap

- `fmt.Sprintf(format, values...)` fills each verb with the next value and
  returns a string.
- `%s` strings, `%d` ints, `%q` a quoted string, `%.1f` a float with one
  decimal, `%%` a percent sign.
- A width such as `%5d` pads with spaces.

### vet errors

Some mistakes compile fine but are almost certainly bugs, like `%d` given a
string. `go vet` catches them, and Gopherlings runs it with the tests. While
vet reports a problem, **every test shows ✗**:

```text
./exercise.go:10:30: fmt.Sprintf format %d has arg service of wrong type string
```

Read it in parts: line 10, the verb `%d`, the value `service`, and its type
`string`. The verbs and values are matched up in order, so the first verb
goes with the first value after the format string.

> **Why is it written like this?** **Go requires** only that the format is a
> string; the compiler doesn't read verbs. **Convention:** every Go project
> runs `go vet`, which does read them, because a wrong verb prints garbage
> like `%!d(string=api)`.

### Padding with zeros

Put a `0` before the width to pad a number with zeros instead of spaces:

```go
fmt.Sprintf("%04d", 42)   // => "0042"
fmt.Sprintf("%03d", 7)    // => "007"
fmt.Sprintf("%03d", 1234) // => "1234" (already wider than 3)
```

## Your task

Fix and finish the bot's messages. The starter does **not pass `go vet`**, so
every test shows ✗ until you finish task 1.

### 1. Fix the scale message

**Goal:** fix the format string in `ScaleMessage` (line 10) so it returns the
message below.

```go
ScaleMessage("api", 2, 5)
// => "scaling api from 2 to 5 replicas"

ScaleMessage("worker", 10, 4)
// => "scaling worker from 10 to 4 replicas"
```

Only the format string (the part in quotes) needs to change.

**Think it through:**
- Match each verb to its value in order. Which value is a string?

### 2. Write a logfmt line

**Goal:** make `LogLine` return `level=`, the level, ` msg=`, the message in
double quotes, ` duration_ms=`, then the duration.

```go
LogLine("info", "deploy finished", 1530)
// => `level=info msg="deploy finished" duration_ms=1530`

LogLine("error", "health check failed", 87)
// => `level=error msg="health check failed" duration_ms=87`
```

*logfmt* is a log format many tools (Grafana Loki, Heroku, Go's `log/slog`)
understand: `key=value` pairs separated by spaces, with text that may contain
spaces in double quotes. The level is not quoted.

**Think it through:**
- Which verb gives the message its quotes?
- Three values means three verbs. What type is each value?

### 3. Format the incident ID

**Goal:** make `IncidentID` return `INC-` followed by `n` padded with zeros to
at least 5 digits.

```go
IncidentID(42)     // => "INC-00042"
IncidentID(7)      // => "INC-00007"
IncidentID(123456) // => "INC-123456"
```

**Think it through:**
- `INC-` is fixed text in the format string. What verb follows it?

## Learn more

- [Package fmt](https://pkg.go.dev/fmt)
- [Command vet: printf check](https://pkg.go.dev/cmd/vet)
- [Brandur: logfmt](https://brandur.org/logfmt)
