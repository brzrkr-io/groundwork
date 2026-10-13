# Variables: fix the broken builders

Someone on your team wrote a few string builders for the log shipper and the
alert bot. One doesn't compile and one has a bug. Time to fix them.

**New here:** reading a compiler error, the `+=` shortcut.

## Learn

### Recap

- Parameters are a function's named inputs. Use them by name, without quotes.
- `name := value` **creates** a variable; `name = value` **changes** one.
- `+` joins strings. Go adds no spaces or separators; you add each one.

### Reading a compiler error

Go checks your code before it runs anything. If something is wrong, it prints
the file, line and column, then a message:

```text
./exercise.go:9:7: no new variables on left side of :=
```

Read it in parts:

| Part | Meaning |
|---|---|
| `./exercise.go` | the file |
| `9` | the line number (your editor shows these on the left) |
| `7` | the column: how many characters along the line |
| `no new variables on left side of :=` | what's wrong |

This message means `:=` tried to create a variable that already exists.
While the code doesn't compile, **every test shows ✗**, because no test can
run. Fix the error and the tests run again.

> **Why is it written like this?** **Go requires** that `:=` creates at least
> one new name. That rule catches typos: if you meant to change a variable
> but made a new one by accident, Go tells you.

### The `+=` shortcut

`x += y` is short for `x = x + y`. It's handy when you build a string step by
step:

```go
msg := "deploy"
msg += " started" // msg is now "deploy started"
```

Note the space inside the quotes before `started`. Without it you'd get
`deploystarted`.

> **Why is it written like this?** **Go requires** the variable to exist
> already, because `+=` changes it. **Your choice:** `+=` or the long form
> `x = x + y`. They do the same thing.

## Your task

Fix the FQDN, log path and alert title helpers.

### 1. Fix the compile error in FQDN

**Goal:** make `FQDN` compile and return the host, a dot, then the domain.

```go
FQDN("web-01", "prod.example.com")
// => "web-01.prod.example.com"

FQDN("db-02", "staging.example.com")
// => "db-02.staging.example.com"
```

The error points at line 9, the line under `TODO(1)`. Only that line needs to
change. A host name written out in full like this is a *fully qualified
domain name* (FQDN).

**Think it through:**
- On line 9, is `name` being created or changed? Look at the line above it.
- Which operator means "change", and which means "create"?

### 2. Fix the log path bug

**Goal:** make `LogPath` return `/var/log/`, the service, a slash, the date,
then `.log`.

```go
LogPath("nginx", "2026-10-11")
// => "/var/log/nginx/2026-10-11.log"

LogPath("api", "2026-01-31")
// => "/var/log/api/2026-01-31.log"
```

It compiles, but the result is wrong. Compare what the test says you returned
with what it wanted.

**Think it through:**
- Which character is missing in your result, and between which two pieces?
- That missing piece is fixed text. How is fixed text written?

### 3. Finish the alert title

**Goal:** in `AlertTitle`, use `+=` to add a space, the service, a colon and a
space, then the summary.

```go
AlertTitle("critical", "api", "disk almost full")
// => "[critical] api: disk almost full"

AlertTitle("warning", "billing", "slow responses")
// => "[warning] billing: slow responses"
```

The line above `TODO(3)` already adds `[`, the severity and `]`. The pieces
you add: fixed text ` ` (one space), the `service` parameter, fixed text
`: ` (colon then space), the `summary` parameter.

**Think it through:**
- The line above `TODO(3)` shows the pattern. What changes?
- Where do the spaces go: inside quotes or outside?

## Learn more

- [Go spec: Short variable declarations](https://go.dev/ref/spec#Short_variable_declarations)
- [Go spec: Assignment statements](https://go.dev/ref/spec#Assignment_statements)
- [Go by Example: Variables](https://gobyexample.com/variables)
