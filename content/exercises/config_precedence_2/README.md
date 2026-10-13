# Config precedence: fix the backup job config

The nightly database backup job reads defaults, a JSON file, environment
variables and flags. A teammate wrote the loader, but it ignores what on-call
asks for. Fix two bugs and fill one gap.

**New here:** `time.ParseDuration`, `strconv.ParseBool`, "set but empty".

## Learn

### Recap

Layers override in this order: defaults < file < environment < flags. Each
layer only overrides what it mentions. `lookup` (like `os.LookupEnv`) returns
a value plus `ok`, which says whether the variable is set. `fs.Visit` only
visits flags that were typed.

### Durations and booleans from strings

Environment values are always strings, so typed settings need parsing. Both
parsers return a value and an error:

| Function | Accepts | Rejects |
|---|---|---|
| `time.ParseDuration(s)` | number + unit (`ns us ms s m h`): `"90s"`, `"5m"`, `"1h30m"` | `"5 minutes"`, `"300"` (no unit) |
| `strconv.ParseBool(s)` | `1 t T TRUE true True 0 f F FALSE false False` | `"yes"`, `"maybe"` |

```go
d, err := time.ParseDuration("2h") // d == 2 * time.Hour, err == nil
on, err := strconv.ParseBool("T")  // on == true
```

> **Why is it written like this?** Go has no exceptions, so a parser that can
> fail returns two values (Go requires you to receive both, or write `_`).
> Checking `if err != nil` right away is convention.

Flags have matching types: `fs.DurationVar` parses `-interval 10m`, and
`fs.BoolVar` makes a switch. `-compress` alone means true; turning it off
needs the `=`: `-compress=false`.

### When empty means something

Sometimes empty is a deliberate choice. `HTTPS_PROXY` sends HTTPS traffic
through a proxy. To turn the file's proxy off for one run:

```sh
HTTPS_PROXY= backup-job
```

The variable is *set*, to `""`, and that must override the file. Code that
checks `v != ""` can't see it. Only `ok` can.

## Your task

All code is in `exercise.go`. `Defaults` gives interval `1h`, compression on,
no proxy. `Source` keys are `"interval"`, `"compress"`, `"proxy"`; values are
`"default"`, `"file"`, `"env"` or `"flag"`.

### 1. Put the layers in the right order

**Goal:** in `Load`, make the flags apply **after** the environment.

Right now `fs.Visit` runs before `s.ApplyEnv`, so an environment variable
beats a flag typed by hand. The code inside each block is fine; only the
order is wrong.

```go
// HTTPS_PROXY=http://env-proxy:3128
Load([]string{"-proxy", "http://debug-proxy:8888"}, lookup)
// Proxy == "http://debug-proxy:8888", Source["proxy"] == "flag"
```

An empty flag also wins: `-proxy=` gives proxy `""` with source `"flag"`.

**Think it through:**
- When two layers set the same thing, which one wins: the first or the last to run?
- Read `Load` top to bottom after the `TODO(1)`. Which block runs last?

### 2. Let HTTPS_PROXY="" disable the proxy

**Goal:** in `ApplyEnv`, make any **set** `HTTPS_PROXY` (empty or not)
override the proxy with source `"env"`.

An unset `HTTPS_PROXY` changes nothing.

```go
s := Defaults()
s.ApplyFile([]byte(`{"proxy": "http://proxy.internal:3128"}`))
s.ApplyEnv(lookup) // HTTPS_PROXY is set to ""
// s.Proxy == "", s.Source["proxy"] == "env"
```

**Think it through:**
- `lookup` returns two values. The starter throws one away with `_`. Which?
- What should the `if` test instead of `v != ""`?

### 3. Parse the interval and compression variables

**Goal:** make `ApplyEnv` also read `BACKUP_INTERVAL` and `BACKUP_COMPRESS`,
each with source `"env"` when set.

| Variable | Setting | Parse with |
|---|---|---|
| `BACKUP_INTERVAL` | `Interval` | `time.ParseDuration` |
| `BACKUP_COMPRESS` | `Compress` | `strconv.ParseBool` |

If parsing fails, return an error whose text contains the variable's name,
for example `invalid BACKUP_INTERVAL "5 minutes": ...`. Add `"strconv"` to
the imports.

```go
// BACKUP_INTERVAL=5m, BACKUP_COMPRESS=false
s := Defaults()
s.ApplyEnv(lookup)
// s.Interval == 5*time.Minute, s.Compress == false
```

**Think it through:**
- Each variable: look it up, parse, check the error, store value and source.
- `ApplyFile` already parses an interval. What can you borrow from it?

## Learn more

- [Package time: ParseDuration](https://pkg.go.dev/time#ParseDuration)
- [Package strconv: ParseBool](https://pkg.go.dev/strconv#ParseBool)
- [Package flag](https://pkg.go.dev/flag)
- [Package net/http: ProxyFromEnvironment](https://pkg.go.dev/net/http#ProxyFromEnvironment)
