# Flags: repeatable and validated flags

During an incident you silence noisy alerts, for example everything with
`team=sre` and `alertname=DiskFull` in `prod`, for two hours. A `silence` tool
needs flags that can repeat and flags that only accept certain values.

## Learn

Recap from *Flags: parsing options with a FlagSet*: create a flag set with
`flag.NewFlagSet(name, flag.ContinueOnError)` and `fs.SetOutput(stderr)`.
`fs.String`, `fs.Duration` and friends return pointers that `fs.Parse` fills
in. When parsing fails, `flag` prints the error and usage text itself, and
`Parse` returns an error.

### Read pointers after Parse, not before

`*p` copies the value the pointer points at **at that moment**. Copy it before
`Parse` and you get the default, and later changes are lost:

```go
author := fs.String("author", "oncall", "who created it")
name := *author   // "oncall", copied too early
fs.Parse([]string{"-author", "pat"})
fmt.Println(name) // still "oncall"
```

### fs.Func: your own flag logic

`fs.Func(name, usage, fn)` defines a flag without a built-in type. Every time
the flag appears on the command line, `Parse` calls your function `fn` with
the text. If `fn` returns an error, `Parse` treats it like any other bad value:

```go
level := "info"
fs.Func("level", "log level", func(s string) error {
	if s != "info" && s != "debug" {
		return errors.New("level must be info or debug")
	}
	level = s
	return nil
})
```

The function is a *closure*, as in *Functions as values*: it updates `level`, a variable
outside it. The variable's starting value is the default. On
`-level trace`, `flag` prints this, then the usage text:

```text
invalid value "trace" for flag -level: level must be info or debug
```

Because `fn` runs once per occurrence, `fs.Func` is also how you build a
*repeatable flag*: `-label team=sre -label tier=web` calls it twice.

`strings.Cut(s, "=")` splits a string at the first `=`. It returns the part
before, the part after, and whether `=` was found:

```go
strings.Cut("team=sre", "=") // "team", "sre", true
strings.Cut("team", "=")     // "team", "", false
```

## Your task

`ParseArgs` in `exercise.go` turns the `silence` arguments into an `Options`
struct.

### 1. Fix the defaults bug

`-for` and `-author` are defined, but their values are always the defaults
(`1h` and `oncall`), whatever you pass. Fix it so the parsed values end up in
`Options`.

```go
ParseArgs([]string{"-for", "30m", "-author", "pat"}, stderr)
// => For: 30 * time.Minute, Author: "pat"
ParseArgs([]string{}, stderr)
// => For: time.Hour, Author: "oncall"
```

### 2. Add a repeatable -label flag

Each `-label` takes `key=value` and adds that pair to `Options.Labels`. Split
at the first `=` only, so the value may contain more `=` signs. When there's
no `=`, or the key or the value is empty, return the error
`label must look like key=value`. With no `-label` flags, `Labels` is empty.

```go
ParseArgs([]string{"-label", "team=sre", "-label", "alertname=DiskFull"}, stderr)
// => Labels: map[string]string{"team": "sre", "alertname": "DiskFull"}

ParseArgs([]string{"-label", "team"}, stderr)
// => error; stderr starts with:
// invalid value "team" for flag -label: label must look like key=value
```

### 3. Add a validated -env flag

`-env` must be exactly `dev`, `staging` or `prod`, and defaults to `dev`.
Anything else returns the error `env must be dev, staging or prod`. Store it in
`Options.Env`.

```go
ParseArgs([]string{"-env", "prod"}, stderr) // => Env: "prod"
ParseArgs([]string{}, stderr)               // => Env: "dev"

ParseArgs([]string{"-env", "production"}, stderr)
// => error; stderr starts with:
// invalid value "production" for flag -env: env must be dev, staging or prod
```

## Learn more

- [pkg.go.dev: FlagSet.Func](https://pkg.go.dev/flag#FlagSet.Func)
- [pkg.go.dev: strings.Cut](https://pkg.go.dev/strings#Cut)
- [Prometheus Alertmanager: Silences](https://prometheus.io/docs/alerting/latest/alertmanager/#silences)
