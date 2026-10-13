# Regular expressions: fix the patterns

A regex that is a little bit wrong is worse than none: it lets bad hostnames
through or leaks secrets into your log platform. Here you fix two broken
patterns and use a regex to rewrite text.

**New here:** `|` (or), `\b` (word boundary), repeating a group,
`ReplaceAllString` with `${1}`.

The starter compiles and runs. Each bug shows up as a failing test.

## Learn

### Recap

As in *Regular expressions: release tags and log lines*: compile once at
package level with `regexp.MustCompile`, check with `MatchString`. Without
`^` and `$` a pattern can match anywhere inside the text. A plain `.` matches
*any* character; `\.` is a real dot.

### Why backticks matter

In a `"double-quoted"` Go string, a backslash starts an escape. Some escapes
don't exist, so the compiler stops you:

```go
var digits = regexp.MustCompile("\d+")
```

```text
./exercise.go:5:35: unknown escape
```

That's file, line 5, column 35, as in *Reading compiler errors*.

Other escapes *do* exist, and those are worse: the code compiles and the
pattern silently changes. In double quotes, `"\t"` is a tab character, not
the two characters `\` and `t` the regexp package needs. Backticks avoid both
problems.

> **Why is it written like this?** Go decides what `\` means in a
> double-quoted string *before* the regexp package sees it: a language rule.
> Backticks turn escapes off, which is why they're the convention for
> patterns.

### Two new pieces of syntax

| Pattern | Matches |
|---|---|
| `a\|b` | `a` or `b`. In a group: `(GET\|POST)` |
| `\b` | a *word boundary*: the edge between a word character (`\w`) and anything else, or the start/end of the text. It matches no characters itself |

So `\bWARN\b` matches "disk WARN" and "WARN:", but not "WARNING" or
"NOWARN".

### Repeating a group

`+` after a group repeats the whole group. `(ab)+` matches "ab", "abab",
"ababab". With a path:

```go
var dirs = regexp.MustCompile(`^(/[a-z]+)+$`)

dirs.MatchString("/var/log") // true: "/var" then "/log"
dirs.MatchString("/var//log") // false: "//" has an empty name
```

### Replacing matches

`ReplaceAllString(src, repl)` returns a copy of `src` with every match
replaced by `repl`. Inside `repl`, `${1}` stands for what capture group 1
matched:

```go
var user = regexp.MustCompile(`(user)=\w+`)

user.ReplaceAllString("host=db-1 user=alice", "${1}=someone")
// => "host=db-1 user=someone"
```

> **Why is it written like this?** You may see `$1` without braces. Prefer
> `${1}` (convention): in `$1x`, Go reads the group name as `1x`, which
> doesn't exist, and silently inserts nothing.

## Your task

Your team ships application logs to a shared log platform. Before a line
leaves the host, it is checked for alerts, its hostname is validated, and its
secrets are removed.

### 1. Fix the alert matcher

**Goal:** fix the `alertWord` pattern so `IsAlert` reports whether a line
contains `ERROR` or `FATAL` as a whole word.

Uppercase only. Right now it never returns `true`. `IsAlert` itself is fine;
change only the pattern.

```go
IsAlert("09:14:03 ERROR disk full on /var") // => true
IsAlert("FATAL: out of memory")             // => true
IsAlert("retrying after ERROR")             // => true
IsAlert("checks passed, ERRORS=0")          // => false (ERRORS is a different word)
IsAlert("dns answer NOERROR")               // => false
IsAlert("level=error msg=timeout")          // => false (lowercase)
```

**Think it through:**
- Look at the quotes around the pattern. What does Go do with `\b` inside
  them?
- Is the pattern itself right once Go stops changing it?

### 2. Fix the hostname check

**Goal:** fix the `internalHost` pattern so `IsInternalHost` accepts only
hostnames like `db-1.internal`.

A valid host is one or more *labels*, each followed by a dot, then
`internal`. A label is one or more lowercase letters, digits or `-`. Nothing
may come before or after.

```go
IsInternalHost("db-1.internal")           // => true
IsInternalHost("api.eu-west-1.internal")  // => true (two labels)
IsInternalHost("db-1xinternal")           // => false (needs a real dot)
IsInternalHost("db-1.internal.evil.com")  // => false (text after)
IsInternalHost("DB-1.internal")           // => false (uppercase)
IsInternalHost("my host.internal")        // => false (space)
IsInternalHost("internal")                // => false (no label)
IsInternalHost(".internal")               // => false (empty label)
IsInternalHost("db..internal")            // => false (empty label)
```

**Think it through:**
- Which `.` in the current pattern should be a real dot?
- What stops extra text before or after the host?
- "Label then dot" repeats. How did the `dirs` example repeat "slash then
  name"?

### 3. Redact secrets

**Goal:** make `Redact(line)` replace the value after `password=`, `token=`
or `secret=` with `REDACTED`.

- The value is everything up to the next space or the end of the line (one
  or more non-space characters).
- The key and the `=` stay; the rest of the line is unchanged.
- Keys are lowercase. A key inside a longer word still counts: `api_token=`
  contains `token=`.

```go
Redact("user=deploy password=hunter2 host=db-1")
// => "user=deploy password=REDACTED host=db-1"

Redact("token=abc123 secret=s3cr3t!")
// => "token=REDACTED secret=REDACTED"

Redact("api_token=xyz789 retries=3")
// => "api_token=REDACTED retries=3"

Redact("connecting to db-1.internal:5432")
// => "connecting to db-1.internal:5432" (unchanged)
```

Two places change: the `secretPair` pattern and the replacement text in
`Redact`.

**Think it through:**
- `secretPair` stops at the `=`. Which table row in *Regular expressions:
  release tags and log lines* means "one non-space character"?
- In the replacement, `${1}` puts the key back. What fixed text comes after
  it?

## Learn more

- [Package regexp: Regexp.ReplaceAllString](https://pkg.go.dev/regexp#Regexp.ReplaceAllString)
- [Package regexp: Regexp.Expand (the `${1}` syntax)](https://pkg.go.dev/regexp#Regexp.Expand)
- [RE2 syntax reference](https://pkg.go.dev/regexp/syntax)
- [Go spec: Rune literals (Go's escape sequences)](https://go.dev/ref/spec#Rune_literals)
