# Formatting: fix the bot messages

The scaling bot's messages are broken: one won't even build, one is missing,
and incident tickets need properly padded IDs.

## Learn

### Recap

`fmt.Sprintf(format, values...)` fills each verb with the next value and
returns a string. `%s` is for strings, `%d` for ints, `%q` for a quoted
string, `%.1f` for a float with one decimal and `%%` for a percent sign. A
width such as `%5d` pads with spaces.

### vet errors

Some mistakes compile fine but are almost certainly bugs, like `%d` given a
string. `go vet` catches them, and Gopherlings runs it with the tests. While
vet reports a problem, **every test shows ✗**:

```text
./exercise.go:9:30: fmt.Sprintf format %d has arg service of wrong type string
```

It points at the line and names the verb and the value that don't match.

### Padding with zeros

Put a `0` before the width to pad a number with zeros instead of spaces:

```go
fmt.Sprintf("%05d", 42)  // => "00042"
fmt.Sprintf("%03d", 7)   // => "007"
fmt.Sprintf("%03d", 1234) // => "1234" (already wider than 3)
```

## Your task

Fix and finish the bot's messages. The starter does **not pass `go vet`**, so
every test shows ✗ until you finish task 1.

### 1. Fix the scale message

`ScaleMessage` has a vet error. Read it and fix the format string so the
function returns:

```go
ScaleMessage("api", 2, 5)
// => "scaling api from 2 to 5 replicas"

ScaleMessage("worker", 10, 4)
// => "scaling worker from 10 to 4 replicas"
```

### 2. Write a logfmt line

*logfmt* is a log format that many tools (Grafana Loki, Heroku, Go's own
`log/slog`) understand: `key=value` pairs separated by spaces, with text that
may contain spaces in double quotes. Write `LogLine` so it returns the level,
the message in double quotes, and the duration in milliseconds:

```go
LogLine("info", "deploy finished", 1530)
// => `level=info msg="deploy finished" duration_ms=1530`

LogLine("error", "health check failed", 87)
// => `level=error msg="health check failed" duration_ms=87`
```

### 3. Format the incident ID

Incident tickets look like `INC-00042`: the text `INC-` and then the number
padded with zeros to at least 5 digits. Write `IncidentID`:

```go
IncidentID(42)     // => "INC-00042"
IncidentID(7)      // => "INC-00007"
IncidentID(123456) // => "INC-123456"
```

## Learn more

- [Package fmt](https://pkg.go.dev/fmt)
- [Command vet: printf check](https://pkg.go.dev/cmd/vet)
- [Brandur: logfmt](https://brandur.org/logfmt)
