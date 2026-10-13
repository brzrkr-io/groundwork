# Regular expressions: fix the patterns

A regex that is a little bit wrong is worse than none: it lets bad hostnames
through or leaks secrets into your log platform. Here you fix two broken
patterns and use a regex to rewrite text.

The starter compiles and runs. Each bug shows up as a failing test.

## Learn

### Recap

`regexp.MustCompile` compiles a pattern once, at package level. `MatchString`
reports whether it matches. Without `^` and `$` a pattern can match anywhere
inside the text, and a plain `.` matches *any* character; write `\.` for a
real dot.

### Why backticks matter

In a `"double-quoted"` Go string, a backslash starts an escape. Some escapes
don't exist, so the compiler stops you:

```go
var digits = regexp.MustCompile("\d+")
```

```text
./exercise.go:5:35: unknown escape
```

The message gives the file, the line (5) and the column (35) where the bad
escape is. The fix is to write the pattern in backticks: `` `\d+` ``.

Other escapes *do* exist in Go strings, and those are worse: the code compiles
and the pattern silently changes. `"\b"` is the *backspace* character, not the
regex `\b`. Write every pattern in backticks and both problems go away.

### Two new pieces of syntax

| Pattern | Matches |
|---|---|
| `a\|b` | `a` or `b`. Inside a group: `(ERROR\|FATAL)` |
| `\b` | a *word boundary*: the edge between a word character (`\w`) and anything else, or the start or end of the text. It matches no characters itself |

So `\bERROR\b` matches "disk ERROR" and "ERROR:", but not "ERRORS" or
"NOERROR".

A group can repeat too: `([a-z]+\.)+` means "one or more lowercase words, each
followed by a dot", so it matches "api." and "api.eu."

### Replacing matches

`ReplaceAllString(src, repl)` returns a copy of `src` with every match
replaced by `repl`. Inside `repl`, `${1}` stands for whatever capture group 1
matched:

```go
var port = regexp.MustCompile(`(port)=\d+`)

port.ReplaceAllString("host=db-1 port=5432", "${1}=XXXX")
// => "host=db-1 port=XXXX"
```

You may also see `$1` without braces. Prefer `${1}`: in `$1x`, Go reads the
group name as `1x`, which doesn't exist, and silently inserts nothing.

## Your task

Your team ships application logs to a shared log platform. Before a line
leaves the host, it is checked for alerts, its hostname is validated, and its
secrets are removed.

### 1. Fix the alert matcher

`IsAlert` should report whether a line contains `ERROR` or `FATAL` as a whole
word (uppercase only). Right now it never returns true. Fix the `alertWord`
pattern; `IsAlert` itself is fine.

```go
IsAlert("09:14:03 ERROR disk full on /var") // => true
IsAlert("FATAL: out of memory")             // => true
IsAlert("checks passed, ERRORS=0")          // => false (ERRORS is a different word)
IsAlert("dns answer NOERROR")               // => false
IsAlert("level=error msg=timeout")          // => false (lowercase)
```

### 2. Fix the hostname check

`IsInternalHost` should accept a hostname only if it is one or more *labels*
followed by `internal`, each label followed by a dot. A label is one or more
lowercase letters, digits or `-`. Nothing may come before or after.

The current `internalHost` pattern accepts hosts it shouldn't. Fix the
pattern.

```go
IsInternalHost("db-1.internal")           // => true
IsInternalHost("api.eu-west-1.internal")  // => true
IsInternalHost("db-1xinternal")           // => false (needs a real dot)
IsInternalHost("db-1.internal.evil.com")  // => false (text after)
IsInternalHost("DB-1.internal")           // => false (uppercase)
IsInternalHost("my host.internal")        // => false (space)
IsInternalHost("internal")                // => false (no label)
IsInternalHost(".internal")               // => false (empty label)
IsInternalHost("db..internal")            // => false (empty label)
```

### 3. Redact secrets

Wherever `password=`, `token=` or `secret=` appears in a line, replace the
value after the `=` with `REDACTED`. The value is everything up to the next
space or the end of the line (one or more non-space characters). The key stays
as it is, and every other part of the line is unchanged. Keys are lowercase.

`secretPair` already matches the key and the `=`. Extend it to match the value
too. `Redact` already calls `secretPair.ReplaceAllString`, which puts the key
back with `${1}` and the `=` after it. Change the replacement so `REDACTED`
follows the `=`.

```go
Redact("user=deploy password=hunter2 host=db-1")
// => "user=deploy password=REDACTED host=db-1"

Redact("token=abc123 secret=s3cr3t!")
// => "token=REDACTED secret=REDACTED"

Redact("api_token=xyz789 retries=3")
// => "api_token=REDACTED retries=3" (contains "token=", so it is redacted too)

Redact("connecting to db-1.internal:5432")
// => "connecting to db-1.internal:5432" (no secrets: unchanged)
```

## Learn more

- [Package regexp: Regexp.ReplaceAllString](https://pkg.go.dev/regexp#Regexp.ReplaceAllString)
- [Package regexp: Regexp.Expand (the `${1}` syntax)](https://pkg.go.dev/regexp#Regexp.Expand)
- [RE2 syntax reference](https://pkg.go.dev/regexp/syntax)
- [Go spec: Rune literals (Go's escape sequences)](https://go.dev/ref/spec#Rune_literals)
