# Config precedence: fix the backup job config

The nightly database backup job reads its settings from defaults, a JSON
file, environment variables and flags. A teammate wrote the loader, but on
call keeps finding it ignores what they ask for. Two bugs, one gap.

## Learn

### Recap

Settings come in layers, each overriding only what it mentions: defaults <
config file < environment variables < flags. `os.LookupEnv` (here, an
injected `lookup` function) reports whether a variable is set at all,
which `os.Getenv` can't. JSON pointer fields are `nil` when a key is missing.
`fs.Visit` calls your function only for the flags that were actually set.

### Durations and booleans from strings

Environment variables are always strings, so typed settings need parsing.
Both parsers return an error for bad input:

| Function | Accepts | Example |
|---|---|---|
| `time.ParseDuration(s)` | a number plus unit: `ns`, `us`, `ms`, `s`, `m`, `h` | `"90s"`, `"5m"`, `"1h30m"` |
| `strconv.ParseBool(s)` | `1 t T TRUE true True 0 f F FALSE false False` | `"false"` |

```go
d, err := time.ParseDuration("5m") // d == 5 * time.Minute
b, err := strconv.ParseBool("0")   // b == false
_, err = time.ParseDuration("5 minutes") // err != nil
```

The flag package has matching flag types. `fs.DurationVar` parses
`-interval 10m`. `fs.BoolVar` makes a switch: `-compress` alone means true,
and to turn it off you write `-compress=false` (with the `=`).

### When empty means something

For most settings, an empty value is just a mistake. For some, it's a
deliberate choice. `HTTPS_PROXY` tells programs to send HTTPS traffic through
a proxy. If the config file sets a proxy, an operator can turn it off for one
run with:

```sh
HTTPS_PROXY= backup-job
```

The variable is set, to the empty string, and that has to override the file.
Code that only checks `v != ""` can't see this, which is exactly the
`os.Getenv` trap.

## Your task

All the code is in `exercise.go`. `Defaults` gives an interval of `1h`,
compression on, and no proxy. The setting keys in `Source` are `"interval"`,
`"compress"` and `"proxy"`, with sources `"default"`, `"file"`, `"env"` or
`"flag"`.

### 1. Put the layers in the right order

`Load` applies the flags **before** the environment, so an environment
variable beats a flag typed by hand. Fix the order so flags come last.

```go
// HTTPS_PROXY=http://env-proxy:3128
Load([]string{"-proxy", "http://debug-proxy:8888"}, lookup)
// Proxy == "http://debug-proxy:8888", Source["proxy"] == "flag"
```

### 2. Let HTTPS_PROXY="" disable the proxy

`ApplyEnv` ignores `HTTPS_PROXY` when it's set to the empty string. Fix it so
that any **set** `HTTPS_PROXY`, empty or not, overrides the proxy with source
`"env"`. An unset `HTTPS_PROXY` changes nothing.

```go
s := Defaults()
s.ApplyFile([]byte(`{"proxy": "http://proxy.internal:3128"}`))
s.ApplyEnv(lookup) // with HTTPS_PROXY set to ""
// s.Proxy == "", s.Source["proxy"] == "env"
```

### 3. Parse the interval and compression variables

`ApplyEnv` doesn't read the other two variables yet. Add them, each with
source `"env"` when set:

| Variable | Setting | Parse with |
|---|---|---|
| `BACKUP_INTERVAL` | `Interval` | `time.ParseDuration` |
| `BACKUP_COMPRESS` | `Compress` | `strconv.ParseBool` |

When parsing fails, return an error that contains the variable's name, for
example `invalid BACKUP_INTERVAL "5 minutes": ...`. Add `"strconv"` to the
imports.

```go
// BACKUP_INTERVAL=5m, BACKUP_COMPRESS=false
s.ApplyEnv(lookup)
// s.Interval == 5*time.Minute, s.Compress == false
```

## Learn more

- [Package time: ParseDuration](https://pkg.go.dev/time#ParseDuration)
- [Package strconv: ParseBool](https://pkg.go.dev/strconv#ParseBool)
- [Package flag](https://pkg.go.dev/flag)
- [Package net/http: ProxyFromEnvironment](https://pkg.go.dev/net/http#ProxyFromEnvironment)
