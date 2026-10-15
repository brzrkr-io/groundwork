# Flags: repeatable and validated flags

During an incident you silence noisy alerts, for example everything with
`team=sre` in `prod`, for two hours. A `silence` tool needs flags that can
repeat and flags that only accept certain values.

**New here:** `fs.Func`, reading pointers at the right time.

## Learn

### Recap

As in *Flags: parsing options with a FlagSet*: `flag.NewFlagSet(name,
flag.ContinueOnError)` plus `fs.SetOutput(stderr)`. `fs.String`,
`fs.Duration` and friends return pointers that `fs.Parse` fills in. On a bad
value, `flag` prints the error and usage text itself, and `Parse` returns an
error.

### Read pointers after Parse, not before

`*p` copies the value **at that moment**. Copy it before `Parse` and you get
the default forever:

```go
color := fs.String("color", "blue", "paint color")
c := *color        // "blue", copied too early
fs.Parse([]string{"-color", "red"})
fmt.Println(c)     // still "blue"
fmt.Println(*color) // "red"
```

> **Why is it written like this?** Go requires this behaviour: `*p` reads the
> value once; it doesn't stay linked. Where you put the line is your choice,
> so put it after `Parse`.

### fs.Func: your own flag logic

`fs.Func(name, usage, fn)` defines a flag with no built-in type. Each time
the flag appears, `Parse` calls `fn` with its text. If `fn` returns an error,
`Parse` treats it as a bad value:

```go
size := "m"
fs.Func("size", "shirt size: s, m or l", func(s string) error {
	if s != "s" && s != "m" && s != "l" {
		return errors.New("size must be s, m or l")
	}
	size = s
	return nil
})
```

On `-size xl`, `flag` prints this line, then the usage text:

```text
invalid value "xl" for flag -size: size must be s, m or l
```

> **Why is it written like this?** `fn` must have exactly the type
> `func(string) error`: Go requires it. Writing the function inline is a
> *closure* (as in *Functions as values*), so it can change `size`, a
> variable outside it. The variable's starting value is the default (author's
> choice: `fs.Func` has no default argument).

Because `fn` runs once per occurrence, `fs.Func` also builds a *repeatable
flag*: `-tag a -tag b` calls it twice.

### strings.Cut

`strings.Cut(s, sep)` splits at the **first** `sep`. It returns the part
before, the part after, and whether `sep` was found:

```go
strings.Cut("host:8080", ":") // "host", "8080", true
strings.Cut("a:b:c", ":")     // "a", "b:c", true
strings.Cut("host", ":")      // "host", "", false
```

## Your task

`ParseArgs` in `exercise.go` turns the `silence` arguments into an `Options`
struct. You'll need `"errors"` and `"strings"` in the imports.

### 1. Fix the defaults bug

**Goal:** make `-for` and `-author` values reach `Options` (today they're always the defaults).

The defaults are `1h` and `oncall`. Find where `opts` is built and compare it
with where `fs.Parse` runs.

```go
ParseArgs([]string{"-for", "30m", "-author", "pat"}, stderr)
// => For: 30 * time.Minute, Author: "pat"
ParseArgs([]string{}, stderr)
// => For: time.Hour, Author: "oncall"
```

**Think it through:**
- When does `*duration` get read, before or after parsing?
- What has to move, and to where?

### 2. Add a repeatable -label flag

**Goal:** each `-label key=value` adds that pair to `Options.Labels`.

Split at the first `=` only, so `query=a=b` gives key `query`, value `a=b`.
If there's no `=`, or the key or the value is empty, return the error
`label must look like key=value`. With no `-label`, `Labels` is empty.

```go
ParseArgs([]string{"-label", "team=sre", "-label", "alertname=DiskFull"}, stderr)
// => Labels: map[string]string{"team": "sre", "alertname": "DiskFull"}

ParseArgs([]string{"-label", "team"}, stderr)
// => error; stderr starts with:
// invalid value "team" for flag -label: label must look like key=value
```

`=sre` and `team=` are bad too.

**Think it through:**
- Where does the map live so the function can add to it and `Options` can use it?
- Which of `strings.Cut`'s three results tells you `=` was missing?

### 3. Add a validated -env flag

**Goal:** `-env` accepts only `dev`, `staging` or `prod`, defaults to `dev`, and goes in `Options.Env`.

Anything else (including `PROD` and the empty string) returns the error
`env must be dev, staging or prod`.

```go
ParseArgs([]string{"-env", "prod"}, stderr) // => Env: "prod"
ParseArgs([]string{}, stderr)               // => Env: "dev"

ParseArgs([]string{"-env", "production"}, stderr)
// => error; stderr starts with:
// invalid value "production" for flag -env: env must be dev, staging or prod
```

**Think it through:**
- Where does the default come from with `fs.Func`?
- A `switch` with one `case` listing several values is one tidy way to check.

## Learn more

- [pkg.go.dev: FlagSet.Func](https://pkg.go.dev/flag#FlagSet.Func)
- [pkg.go.dev: strings.Cut](https://pkg.go.dev/strings#Cut)
- [Prometheus Alertmanager: Silences](https://prometheus.io/docs/alerting/latest/alertmanager/#silences)
