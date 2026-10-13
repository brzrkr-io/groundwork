# Config precedence: defaults, file, env and flags

Real services read settings from several places at once. You'll merge four
layers in the right order and record where each value came from, so "why is
it listening on 9000?" has an answer.

**New here:** `os.LookupEnv`, JSON pointer fields, `flag.FlagSet.Visit`.

## Learn

### The precedence ladder

Each layer overrides the one above it:

| Layer | Who sets it | Example |
|---|---|---|
| 1. defaults | the developer, in code | port `8080` |
| 2. config file | ops, mounted into the container | `{"port": 9000}` |
| 3. environment | Kubernetes, systemd, CI | `APP_PORT=9100` |
| 4. flags | a human, right now | `-port 9200` |

The key rule: **a layer only overrides the settings it actually mentions.**
A file with just `log_level` leaves the port alone.

So every layer needs a way to tell "missing" apart from "zero" or "empty".

### Environment: `LookupEnv`

`os.Getenv` returns `""` both when a variable is unset and when it's set to
`""`. `os.LookupEnv` returns a second value, `ok`, that says whether it's set:

```go
v, ok := os.LookupEnv("HOME")
// unset:      v == "", ok == false
// HOME="":    v == "", ok == true
```

Here the lookup function is a *parameter* called `lookup`, as in *Fakes and
dependency injection*. Real code passes `os.LookupEnv`; tests pass a fake
backed by a map. You call it the same way: `lookup("SOME_NAME")`.

Environment values are always strings. `strconv.Atoi` turns `"42"` into `42`,
or returns an error.

> **Why is it written like this?** `if v, ok := lookup("X"); ok { … }` is one
> line in two parts. Before the `;` you create `v` and `ok`; after it is the
> condition. Go allows this short statement (Go requires the `;`). The names
> `v` and `ok` are convention, not rules.

### JSON: pointer fields

When a key is missing, `json.Unmarshal` leaves the field at its zero value.
With a plain `int`, `{"port": 0}` and `{}` look the same. A *pointer* field
(`*int`) stays `nil` when the key is absent:

```go
type cache struct {
	Size *int `json:"size"`
}
// {}            -> Size == nil
// {"size": 0}   -> Size != nil, *Size == 0
```

`*c.Size` (star in front) reads the value the pointer points at. Only do that
after checking it isn't `nil`.

### Flags: `Visit`

After `fs.Parse(args)`, a flag nobody typed still holds its default value.
`fs.Visit` calls your function once for each flag that was **actually typed**,
and skips the rest:

```go
fs.Visit(func(f *flag.Flag) {
	switch f.Name {
	case "timeout":
		fmt.Println("you set -timeout")
	}
})
```

`f.Name` is the flag's name without the dash. (`fs.VisitAll` visits every
flag, typed or not. Don't use it here.)

> **Why is it written like this?** `Visit` wants a function, so you write one
> inline: `func(f *flag.Flag) { … }`. Go requires that exact parameter type;
> the name `f` is your choice. The function can change variables outside it,
> like `s`.

### Recording the source

`Settings.Source` is a map from setting key to the layer that set it. Each
time you change a setting, change its entry too. Keys: `"port"`,
`"log_level"`, `"region"`. Values: `"default"`, `"file"`, `"env"`, `"flag"`.

## Your task

The `ingest` service needs its settings loaded properly. `Defaults`, the
`fileSettings` type and most of `Load` are written for you.

### 1. Apply the config file

**Goal:** finish `ApplyFile` so `log_level` and `region` from the JSON override
the settings, with source `"file"`.

Only override a setting when its key is present. The `f.Port` block that's
already there is the pattern. Invalid JSON returns an error (already done).

```go
s := Defaults()
s.ApplyFile([]byte(`{"log_level": "debug"}`))
// s.Port == 8080,        s.Source["port"] == "default"
// s.LogLevel == "debug", s.Source["log_level"] == "file"
```

`{"port": 0}` still counts as present: port `0`, source `"file"`.

**Think it through:**
- Which fields of `fileSettings` are you handling, and how do you know one is missing?
- Which two things change for each present key?

### 2. Apply the environment

**Goal:** finish `ApplyEnv` so each variable that `lookup` reports as set
overrides its setting, with source `"env"`.

| Variable | Setting | Source key |
|---|---|---|
| `APP_PORT` | `Port`, converted with `strconv.Atoi` | `"port"` |
| `APP_LOG_LEVEL` | `LogLevel` | `"log_level"` |
| `APP_REGION` | `Region` | `"region"` |

Unset variables change nothing. If `APP_PORT` isn't a whole number (`"eighty"`,
`"80a"`), return an error whose text contains `APP_PORT`, for example
`invalid APP_PORT "eighty": ...`. Return `nil` when all went well. Add `"fmt"`
and `"strconv"` to the imports.

```go
// lookup knows APP_PORT=9100 and APP_REGION=eu-west-1
s := Defaults()
s.ApplyEnv(lookup)
// s.Port == 9100, s.Region == "eu-west-1", s.LogLevel == "info"
```

**Think it through:**
- Which of `lookup`'s two results tells you "set"?
- The variable name is fixed text in quotes; the value comes back from `lookup`.

### 3. Let flags have the final say

**Goal:** at the `TODO(3)` in `Load`, use `fs.Visit` to apply every flag that
was typed, with source `"flag"`.

`Load` already defines `-config`, `-port`, `-log-level` and `-region`, then
applies defaults, the file and the environment. The parsed flag values sit in
`cli.port`, `cli.logLevel` and `cli.region`.

```go
Load([]string{"-config", "ingest.json", "-port", "9200"}, lookup)
// Port 9200 from "flag"; everything else from the file, env or defaults
```

A typed flag wins even when its value equals the default: `-log-level info`
gives `"info"` with source `"flag"`. `-port 0` gives port `0` from `"flag"`.

**Think it through:**
- Flag *names* (`log-level`, dash) and source *keys* (`log_level`, underscore) differ.
- Where does each typed value live: in `f`, or in `cli`?

## Learn more

- [Package os: LookupEnv](https://pkg.go.dev/os#LookupEnv)
- [Package flag: FlagSet.Visit](https://pkg.go.dev/flag#FlagSet.Visit)
- [The Twelve-Factor App: Config](https://12factor.net/config)
- [Kubernetes: Define environment variables for a container](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)
