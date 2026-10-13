# Regular expressions: release tags and log lines

Release tags, hostnames, log lines: ops work is full of text that follows a
pattern. A *regular expression* (regex) describes that pattern once, and Go's
`regexp` package checks and picks apart any string against it.

## Learn

### What a regular expression is

A regular expression is a small pattern language. Most characters match
themselves: the pattern `ERROR` matches the text "ERROR". A few characters are
special:

| Pattern | Matches |
|---|---|
| `.` | any single character |
| `\d` | one digit, `0`–`9` |
| `\w` | one letter, digit or `_` |
| `\S` | one character that is not a space or tab |
| `[a-z]` | one character from the set: here, a lowercase letter. `[a-z0-9-]` adds digits and `-` |
| `x+` | one or more `x` |
| `x*` | zero or more `x` |
| `x?` | zero or one `x` (optional) |
| `x{3}` | exactly three `x` |
| `^` | the start of the text |
| `$` | the end of the text |
| `\.` | a literal dot (a plain `.` means "any character") |
| `( )` | a *capture group*: remember what matched inside |
| `(?P<name> )` | a capture group with a name |

So `\d+` means "one or more digits", and `^\d{3}$` means "exactly three
digits and nothing else".

Without `^` and `$`, a pattern may match *anywhere* inside the text: `\d+`
matches "port 8080" because "8080" is in there. Add `^...$` when the whole
string must fit.

### Write patterns in backticks

Patterns are full of backslashes. In a normal `"double-quoted"` Go string, a
backslash starts an escape such as `\n`, so you would have to write `"\\d+"`.
A *raw string literal* in backticks keeps every character exactly as typed:

```go
"\\d+\\.\\d+" // double quotes: every backslash doubled
`\d+\.\d+`    // backticks: what you see is the pattern
```

Always use backticks for patterns.

### Compile once, use many times

```go
var port = regexp.MustCompile(`^\d+$`)

port.MatchString("8080") // true
port.MatchString("80a")  // false
```

`regexp.MustCompile` turns the pattern into a `*regexp.Regexp`. Declare it at
package level so it is compiled once, when the program starts. If the pattern
is invalid, `MustCompile` panics. For a pattern written in your code that is
what you want: the program crashes straight away instead of misbehaving later.

`MatchString` reports whether the pattern matches.

### Pull out the parts

`FindStringSubmatch` returns a `[]string`: index 0 holds the whole match, then
one entry per capture group, in order. It returns `nil` when there is no
match.

```go
var hostPort = regexp.MustCompile(`^([a-z0-9.-]+):(\d+)$`)

m := hostPort.FindStringSubmatch("db-1.internal:5432")
// m[0] = "db-1.internal:5432"
// m[1] = "db-1.internal"
// m[2] = "5432"

hostPort.FindStringSubmatch("no port here") // nil
```

The parts are always strings. Turn digits into an `int` with `strconv.Atoi`,
as in *Type conversion*.

Counting groups gets error-prone in long patterns. Name them with
`(?P<name>...)` and look up the index with `SubexpIndex`:

```go
var hostPort = regexp.MustCompile(`^(?P<host>[a-z0-9.-]+):(?P<port>\d+)$`)

m := hostPort.FindStringSubmatch("db-1.internal:5432")
host := m[hostPort.SubexpIndex("host")] // "db-1.internal"
```

A long pattern can be split into pieces joined with `+`, one piece per line.

Go's regexp engine is called *RE2*. It skips a few features of other regex
engines (such as backreferences), and in return a match always runs in time
proportional to the length of the input, so a bad pattern can't hang your
service.

## Your task

Your deploy tooling reads release tags and application logs. Teach it to
recognise both.

### 1. Recognise release tags

A release tag is a lowercase `v`, then three numbers separated by dots, with
nothing before or after. Each number is one or more digits.

In `exercise.go`, replace the placeholder pattern `^v$` in `semverTag` with
the real pattern. `IsSemverTag` already returns `semverTag.MatchString(tag)`,
so it starts working as soon as the pattern is right.

```go
IsSemverTag("v1.4.2")      // => true
IsSemverTag("v10.0.15")    // => true
IsSemverTag("1.4.2")       // => false (no "v")
IsSemverTag("v1.4")        // => false (only two numbers)
IsSemverTag("v1.4.2-rc.1") // => false (text after the tag)
IsSemverTag("xv1.4.2")     // => false (text before the tag)
IsSemverTag("v1x4x2")      // => false (the separators must be dots)
IsSemverTag("")            // => false
```

### 2. Parse release tags

Add a capture group around each of the three numbers in `semverTag`.
`MatchString` ignores groups, so task 1 keeps passing.

Then implement `ParseSemver`. Use `FindStringSubmatch`. If it returns `nil`,
return `Version{}, false`. Otherwise convert the three groups with
`strconv.Atoi` (add `"strconv"` to the imports) and return the `Version` and
`true`.

```go
ParseSemver("v1.4.2")   // => Version{Major: 1, Minor: 4, Patch: 2}, true
ParseSemver("v10.0.15") // => Version{Major: 10, Minor: 0, Patch: 15}, true
ParseSemver("latest")   // => Version{}, false
```

### 3. Parse log lines

Your services log lines like this:

```text
2026-10-11T09:14:03Z ERROR [payments] connection refused
```

The pattern `logLine` is already written for you. Piece by piece:

| Piece | Matches |
|---|---|
| `^(?P<time>\S+) ` | the timestamp: everything up to the first space |
| `(?P<level>[A-Z]+) ` | the level: uppercase letters, then a space |
| `\[(?P<service>[a-z0-9-]+)\] ` | the service in square brackets (`\[` is a literal `[`), then a space |
| `(?P<message>.*)$` | the message: everything else to the end |

Finish `ParseLogLine`. It already calls `logLine.FindStringSubmatch(line)`.
Fill each `LogLine` field from the group with the same name, using
`logLine.SubexpIndex("time")` and so on.

```go
ParseLogLine("2026-10-11T09:14:03Z ERROR [payments] connection refused")
// => LogLine{Time: "2026-10-11T09:14:03Z", Level: "ERROR",
//            Service: "payments", Message: "connection refused"}, nil
```

When the line does not match, return an empty `LogLine` and the error
`not a log line: "<line>"`, with the line in double quotes. Use
`fmt.Errorf("not a log line: %q", line)`.

```go
ParseLogLine("hello world")
// => LogLine{}, error `not a log line: "hello world"`

ParseLogLine("2026-10-11T09:14:03Z error [payments] connection refused")
// => error, because the level must be uppercase
```

## Learn more

- [Go by Example: Regular Expressions](https://gobyexample.com/regular-expressions)
- [Package regexp](https://pkg.go.dev/regexp)
- [RE2 syntax reference](https://pkg.go.dev/regexp/syntax)
- [Go spec: String literals](https://go.dev/ref/spec#String_literals)
