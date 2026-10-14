# Regular expressions: release tags and log lines

Release tags, hostnames, log lines: ops work is full of text that follows a
pattern. A *regular expression* (regex) describes that pattern once, and Go's
`regexp` package checks and picks apart any string against it.

**New here:** regex syntax, `regexp.MustCompile`, `MatchString`,
`FindStringSubmatch`, named groups and `SubexpIndex`. (Raw strings in
backticks are from *Formatting with fmt*.)

## Learn

### What a regular expression is

A regex is a small pattern language. Most characters match themselves: the
pattern `ERROR` matches the text "ERROR". A few are special:

| Pattern | Matches |
|---|---|
| `.` | any single character |
| `\d` | one digit, `0`–`9` |
| `\w` | one letter, digit or `_` |
| `\S` | one character that is not a space or tab |
| `[a-z]` | one character from the set. `[a-z0-9-]` adds digits and `-` |
| `x+` | one or more `x` |
| `x*` | zero or more `x` |
| `x?` | zero or one `x` (optional) |
| `x{3}` | exactly three `x` |
| `^` / `$` | the start / the end of the text |
| `\.` | a literal dot (plain `.` means "any character") |
| `( )` | a *capture group*: remember what matched inside |
| `(?P<name> )` | a capture group with a name |

So `^\d{3}-\d{4}$` means "three digits, a dash, four digits, and nothing
else".

Without `^` and `$`, a pattern may match *anywhere*: `\d+` matches
"port 8080" because "8080" is in there. Add `^...$` when the whole string
must fit.

### Write patterns in backticks

In a `"double-quoted"` Go string, a backslash starts an escape such as `\n`,
so you'd have to double every backslash. A *raw string literal* in backticks
keeps every character as typed:

```go
"\\d+-\\d+" // double quotes: every backslash doubled
`\d+-\d+`   // backticks: what you see is the pattern
```

> **Why is it written like this?** Both are valid Go. Backticks for patterns
> are a strong convention because they avoid the doubling.

### Compile once, use many times

```go
var orderID = regexp.MustCompile(`^ORD-\d+$`)

orderID.MatchString("ORD-1042") // true
orderID.MatchString("ord-1042") // false
```

`regexp.MustCompile` turns the pattern into a `*regexp.Regexp`.
`MatchString` reports whether the text matches.

> **Why is it written like this?** Declaring it at package level (`var`
> outside any function) is convention: it compiles once, at start-up.
> `MustCompile` panics on a bad pattern, which is what you want for a pattern
> written in your code: the program crashes at once instead of misbehaving
> later.

### Pull out the parts

`FindStringSubmatch` returns a `[]string`: index 0 is the whole match, then
one entry per capture group, left to right. It returns `nil` when there is no
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

### Named groups

Counting groups gets error-prone in long patterns. Name them with
`(?P<name>...)` and look up the index with `SubexpIndex`:

```go
var hostPort = regexp.MustCompile(`^(?P<host>[a-z0-9.-]+):(?P<port>\d+)$`)

m := hostPort.FindStringSubmatch("db-1.internal:5432")
port := m[hostPort.SubexpIndex("port")] // "5432"
```

A long pattern can be split into pieces joined with `+`, one per line.

Go's engine (*RE2*) skips a few features of other engines, such as
backreferences. In return, a match always runs in time proportional to the
input, so a bad pattern can't hang your service.

## Your task

Your deploy tooling reads release tags and application logs. Teach it to
recognise both.

### 1. Recognise release tags

**Goal:** replace the placeholder `^v$` in `semverTag` so it matches release
tags like `v1.4.2`.

A release tag is: a lowercase `v`, then three numbers separated by dots,
nothing before or after. Each number is one or more digits. `IsSemverTag`
already calls `semverTag.MatchString(tag)`, so only the pattern changes.

```go
IsSemverTag("v1.4.2")      // => true
IsSemverTag("v10.0.15")    // => true
IsSemverTag("1.4.2")       // => false (no "v")
IsSemverTag("v1.4")        // => false (only two numbers)
IsSemverTag("v1.4.x")      // => false (not a number)
IsSemverTag("v1.4.2-rc.1") // => false (text after the tag)
IsSemverTag("xv1.4.2")     // => false (text before the tag)
IsSemverTag("v1x4x2")      // => false (the separators must be dots)
IsSemverTag("")            // => false
```

**Think it through:**
- Write the tag's pieces in order (start, `v`, number, dot, …, end), then
  find each one in the table.
- A plain `.` would accept `v1x4x2`. Which table row is a real dot?

### 2. Parse release tags

**Goal:** `ParseSemver(tag)` returns the three numbers as a `Version` and
`true`, or `Version{}` and `false` when `tag` isn't a release tag.

First add a capture group around each number in `semverTag`. `MatchString`
ignores groups, so task 1 keeps passing. Then use `FindStringSubmatch` and
`strconv.Atoi` (add `"strconv"` to the imports).

```go
ParseSemver("v1.4.2")   // => Version{Major: 1, Minor: 4, Patch: 2}, true
ParseSemver("v10.0.15") // => Version{Major: 10, Minor: 0, Patch: 15}, true
ParseSemver("latest")   // => Version{}, false
```

**Think it through:**
- What does `FindStringSubmatch` return for `"latest"`? Check that first.
- Which indexes hold the three numbers? (Index 0 is the whole tag.)
- `strconv.Atoi` returns two values. What do you do with the error?

### 3. Parse log lines

**Goal:** finish `ParseLogLine` so it fills each `LogLine` field from its
named group, or returns an error when the line doesn't match.

Your services log lines like this:

```text
2026-10-11T09:14:03Z ERROR [payments] connection refused
```

The pattern `logLine` is already written. Piece by piece:

| Piece | Matches |
|---|---|
| `^(?P<time>\S+) ` | the timestamp: everything up to the first space |
| `(?P<level>[A-Z]+) ` | the level: uppercase letters, then a space |
| `\[(?P<service>[a-z0-9-]+)\] ` | the service in square brackets (`\[` is a literal `[`) |
| `(?P<message>.*)$` | the message: everything else to the end |

Group names `time`, `level`, `service`, `message` fill the fields `Time`,
`Level`, `Service`, `Message`.

```go
ParseLogLine("2026-10-11T09:14:03Z ERROR [payments] connection refused")
// => LogLine{Time: "2026-10-11T09:14:03Z", Level: "ERROR",
//            Service: "payments", Message: "connection refused"}, nil
```

When the line doesn't match, return an empty `LogLine` and the error
`not a log line: "<line>"`: that fixed text, then the line in double quotes.

```go
ParseLogLine("hello world")
// => LogLine{}, error `not a log line: "hello world"`

ParseLogLine("2026-10-11T09:14:03Z error [payments] connection refused")
// => error: the level must be uppercase
```

**Think it through:**
- The starter already handles `m == nil` but with the wrong message. Which
  `fmt` verb puts double quotes around a string?
- Look at the *Named groups* example: how do you get one field's text out of
  `m`?

## Learn more

- [Go by Example: Regular Expressions](https://gobyexample.com/regular-expressions)
- [Package regexp](https://pkg.go.dev/regexp)
- [RE2 syntax reference](https://pkg.go.dev/regexp/syntax)
- [Go spec: String literals](https://go.dev/ref/spec#String_literals)
