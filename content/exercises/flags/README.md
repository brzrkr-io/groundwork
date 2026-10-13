# Flags: parsing options with a FlagSet

Positional arguments get confusing past two or three: was it
`deploy api 3 30s` or `deploy api 30s 3`? Real tools use named *flags*
instead: `deploy -service api -replicas 3 -timeout 30s web-1 web-2`. Go's
`flag` package parses them for you.

## Learn

### A flag set

A *flag set* holds the flags your tool understands. Create one with
`flag.NewFlagSet`, and send its messages to stderr:

```go
fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
fs.SetOutput(stderr)
```

`"deploy"` is the tool name used in messages. `flag.ContinueOnError` means
"on a bad flag, return an error". The `flag` package also has a global flag
set (`flag.String`, `flag.Parse`), but that one calls `os.Exit` on a bad flag
and always reads `os.Args`, which makes it hard to test. Your own flag set has
neither problem.

### Defining flags

Each definition takes a name, a *default* value (used when the flag is not
given) and a help text, and returns a **pointer**:

```go
service := fs.String("service", "", "service to deploy")
replicas := fs.Int("replicas", 1, "number of replicas")
dryRun := fs.Bool("dry-run", false, "print the plan, change nothing")
timeout := fs.Duration("timeout", 30*time.Second, "how long to wait")
```

The pointer is how `Parse` hands you the value later: it writes into the
variable the pointer points at. Read the value with `*`, as in *Pointers*:
`*replicas`. Only do that **after** parsing. `fs.Duration` accepts the same
text as `time.ParseDuration`, such as `90s` or `5m`.

### Parsing

```go
err := fs.Parse(args)
hosts := fs.Args() // the arguments left over after the flags
```

`Parse` wants the arguments **without** the program name, which is
`os.Args[1:]`. From now on `Run` receives exactly that, and `main` becomes:

```go
func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
```

The rules `Parse` follows:

| You type | Meaning |
|---|---|
| `-replicas 3`, `-replicas=3` | set a flag |
| `--replicas 3` | same thing: one or two dashes both work |
| `-dry-run` | a bool flag on its own means `true` |
| `-dry-run=false` | bool flags need `=` to take a value |
| `web-1 -dry-run` | parsing **stops** at the first non-flag argument, so `-dry-run` ends up in `fs.Args()` |

### When parsing fails

`flag` prints its own messages to the flag set's output:

- An unknown flag such as `-bogus`, or a bad value such as `-replicas abc`:
  it prints the error (`flag provided but not defined: -bogus` or
  `invalid value "abc" for flag -replicas: parse error`), then the *usage*
  text, and `Parse` returns an error.
- `-h`, `-help` or `--help`: it prints the usage text and returns the special
  error `flag.ErrHelp`.

The usage text starts with `Usage of deploy:` and lists every flag. You can
print it yourself by calling `fs.Usage()`. Because `flag` has already printed
the message, your code just returns the error and doesn't print it again.

Asking for help isn't a failure, so it gets exit code `0`. Every other parse
error is a usage error, exit code `2`. `errors.Is` tells them apart:

```go
if errors.Is(err, flag.ErrHelp) {
	return 0
}
```

## Your task

Your team's `deploy` tool has a `ParseArgs` function that turns arguments into
an `Options` struct, and a `Run` function that uses it. The `-service` flag is
already defined as an example.

### 1. Define the remaining flags

In `ParseArgs`, define these flags and fill in every field of `Options`:

| Flag | Type | Default | Field |
|---|---|---|---|
| `-replicas` | int | `1` | `Replicas` |
| `-dry-run` | bool | `false` | `DryRun` |
| `-timeout` | duration | `30s` | `Timeout` |

`Hosts` is the leftover arguments, `fs.Args()`.

```go
ParseArgs([]string{"-service", "api", "-replicas=3", "-dry-run", "web-1", "web-2"}, stderr)
// => Options{Service: "api", Replicas: 3, DryRun: true, Timeout: 30 * time.Second,
//            Hosts: []string{"web-1", "web-2"}}, nil
```

### 2. Require -service

After parsing, if `Service` is empty, treat it like one of `flag`'s own
errors: print this line to stderr, then call `fs.Usage()`, and return an empty
`Options` with a non-nil error (`errors.New` is fine; add `"errors"` to the
imports).

```text
ParseArgs([]string{"-replicas", "2"}, stderr) // => Options{}, error
stderr:
deploy: -service is required
Usage of deploy:
  ...the flag list...
```

`-service=` (set to empty) counts as missing too.

### 3. Return the right exit code

`Run` already prints the summary, but always returns `1` when `ParseArgs`
fails. Return `0` when the error is `flag.ErrHelp`, and `2` for any other
error. Don't print anything else: the message is already on stderr. On success
`Run` prints one line to stdout and returns `0`:

```text
Run([]string{"-service", "api", "-replicas", "3", "web-1", "web-2"}, ...) // => 0
stdout: deploy api: replicas=3 timeout=30s dry-run=false hosts=web-1,web-2

Run([]string{"-h"}, ...)     // => 0, usage text on stderr
Run([]string{"-bogus"}, ...) // => 2, error and usage text on stderr
```

A `time.Duration` prints in its own style: `30s`, `5m0s`, `1m30s`. With no
hosts, the line ends in `hosts=`.

## Learn more

- [pkg.go.dev: flag](https://pkg.go.dev/flag)
- [Go by Example: Command-Line Flags](https://gobyexample.com/command-line-flags)
- [pkg.go.dev: time.ParseDuration](https://pkg.go.dev/time#ParseDuration)
