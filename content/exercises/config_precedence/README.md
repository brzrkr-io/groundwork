# Config precedence: defaults, file, env and flags

Real services read their settings from several places at once: built-in
defaults, a config file, environment variables and command-line flags. In
this exercise you merge all four in the right order, and record where each
final value came from, so "why is it listening on 9000?" has an answer.

## Learn

### The precedence ladder

Each layer overrides the one before it:

| Layer | Who sets it | Example |
|---|---|---|
| 1. defaults | the developer, in code | port `8080` |
| 2. config file | ops, mounted into the container | `{"port": 9000}` |
| 3. environment | Kubernetes, systemd, CI | `APP_PORT=9100` |
| 4. flags | a human, right now | `-port 9200` |

The order runs from "most general" to "most specific". The image ships with
sane defaults. A team mounts a shared file. The deployment tweaks one value
per environment. Someone debugging at 3am types a flag, and that flag must
win. The "twelve-factor app" guidelines popularised this style.

The key rule: **a layer only overrides the settings it actually mentions.**
A config file with just `log_level` must leave the port alone.

### "Not set" versus "empty" or "zero"

That rule means each layer must tell "missing" apart from "zero". Each layer
has its own way.

**Environment variables.** `os.Getenv("APP_REGION")` returns `""` both when
the variable is unset and when it's set to the empty string. `os.LookupEnv`
also tells you whether it was set:

```go
v, ok := os.LookupEnv("APP_REGION")
// unset:           v == "", ok == false
// APP_REGION="":   v == "", ok == true
```

To keep tests away from the real environment, the code here takes the lookup
function as a parameter, as in *Fakes and dependency injection*. `main`
passes `os.LookupEnv` itself. Tests pass a closure over a map:

```go
env := map[string]string{"APP_REGION": "eu-west-1"}
lookup := func(key string) (string, bool) {
	v, ok := env[key]
	return v, ok
}
```

Environment values are always strings. Turn a number into an `int` with
`strconv.Atoi`, as in earlier exercises.

**JSON files.** When a key is missing, `json.Unmarshal` leaves the field at
its zero value. So with a plain `int` field, `{"port": 0}` and `{}` look the
same. Make the field a *pointer* instead. It stays `nil` when the key is
absent:

```go
type fileSettings struct {
	Port *int `json:"port"`
}
// {}            -> Port == nil
// {"port": 0}   -> *Port == 0
```

**Flags.** After `fs.Parse(args)`, a flag that wasn't given still holds its
default value. `fs.Visit` calls your function once for each flag that was
**actually set** on the command line, and skips the rest:

```go
fs.Visit(func(f *flag.Flag) {
	switch f.Name {
	case "port":
		s.Port = cli.port // the variable fs.IntVar wrote into
	}
})
```

`f.Name` is the flag's name without the dash. (`fs.VisitAll` visits every
flag, set or not. Don't use it here.)

### Recording the source

`Settings.Source` maps each setting key (`"port"`, `"log_level"`,
`"region"`) to the layer its value came from: `"default"`, `"file"`, `"env"`
or `"flag"`. Every time you override a value, update its source too:

```go
s.Port = *f.Port
s.Source["port"] = "file"
```

## Your task

The `ingest` service needs its settings loaded properly. `Defaults` and the
`fileSettings` type are written for you, and so is most of `Load`.

### 1. Apply the config file

Complete `ApplyFile`. It already decodes the JSON and applies `port`. Apply
`log_level` and `region` the same way: only when present, setting the source
to `"file"`.

```go
s := Defaults()
s.ApplyFile([]byte(`{"log_level": "debug"}`))
// s.Port == 8080,     s.Source["port"] == "default"
// s.LogLevel == "debug", s.Source["log_level"] == "file"
```

`{"port": 0}` sets the port to `0`, with source `"file"`. Invalid JSON
returns an error (the starter already does this).

### 2. Apply the environment

Complete `ApplyEnv`. For each variable that `lookup` reports as set, override
the setting and set its source to `"env"`:

| Variable | Setting |
|---|---|
| `APP_PORT` | `Port`, converted with `strconv.Atoi` |
| `APP_LOG_LEVEL` | `LogLevel` |
| `APP_REGION` | `Region` |

If `APP_PORT` isn't a whole number, return an error that contains
`APP_PORT`, for example `invalid APP_PORT "eighty": ...`. Add `"fmt"` and
`"strconv"` to the imports.

```go
s := Defaults()
s.ApplyEnv(lookup) // with APP_PORT=9100 and APP_REGION=eu-west-1
// s.Port == 9100, s.Region == "eu-west-1", s.LogLevel == "info"
```

### 3. Let flags have the final say

`Load(args, lookup)` defines the flags `-config`, `-port`, `-log-level` and
`-region`, then applies defaults, the file named by `-config` (if any) and
the environment. `fs.StringVar` and `fs.IntVar` store the parsed flag values
in the `cli` struct (`cli.port`, `cli.logLevel`, `cli.region`). Finish
`Load`: after the environment, use `fs.Visit` to apply each flag that was
set, with source `"flag"`. Note that the flag is named `log-level` while the
source key is `log_level`.

```go
Load([]string{"-config", "ingest.json", "-port", "9200"}, lookup)
// Port 9200 from "flag", everything else from the file, env or defaults
```

A flag wins even when its value equals the default: `-log-level info` gives
log level `"info"` with source `"flag"`.

## Learn more

- [Package os: LookupEnv](https://pkg.go.dev/os#LookupEnv)
- [Package flag: FlagSet.Visit](https://pkg.go.dev/flag#FlagSet.Visit)
- [The Twelve-Factor App: Config](https://12factor.net/config)
- [Kubernetes: Define environment variables for a container](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
