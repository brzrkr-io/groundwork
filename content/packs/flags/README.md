# Flags: parsing options with a FlagSet

Positional arguments get confusing fast: was it `deploy api 3 30s` or
`deploy api 30s 3`? Real tools use named *flags* instead:
`deploy -service api -replicas 3 web-1`. Go's `flag` package parses them.

**New here:** `flag.NewFlagSet`, defining and parsing flags, `flag.ErrHelp`.

## Learn

### A flag set

A *flag set* holds the flags your tool understands:

```go
fs := flag.NewFlagSet("backup", flag.ContinueOnError)
fs.SetOutput(stderr)
```

`"backup"` is the tool name used in messages. `flag.ContinueOnError` means
"on a bad flag, return an error". `SetOutput` sends flag's messages to stderr.

> **Why is it written like this?** The `flag` package also has a global set
> (`flag.String`, `flag.Parse`), but it calls `os.Exit` and reads `os.Args`,
> which makes testing hard. Your own flag set is the author's choice for
> testability. `fs` is a conventional short name.

### Defining flags

Each definition takes a **name**, a **default** (used when the flag is not
given) and a **help text**, and returns a pointer:

```go
bucket := fs.String("bucket", "backups", "where to upload")
keep := fs.Int("keep", 7, "how many copies to keep")
verbose := fs.Bool("verbose", false, "print every file")
every := fs.Duration("every", time.Hour, "how often to run")
```

Read the value with `*`, as in *Pointers*: `*keep`. Only do that **after**
parsing. `fs.Duration` accepts text such as `90s` or `5m`.

> **Why is it written like this?** Go requires all three arguments, in that
> order. The name has no dash: `"keep"`, not `"-keep"`. The pointer is
> required by the API: `Parse` writes the value through it later. Naming the
> variable after the flag is convention; Go doesn't care.

### Parsing

```go
err := fs.Parse(args)
rest := fs.Args() // the arguments left over after the flags
```

`Parse` wants the arguments **without** the program name. From now on `Run`
receives `os.Args[1:]`, and `main` is
`os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))`.

| You type | Meaning |
|---|---|
| `-keep 3`, `-keep=3` | set a flag |
| `--keep 3` | same: one or two dashes both work |
| `-verbose` | a bool flag on its own means `true` |
| `-verbose=false` | bool flags need `=` to take a value |
| `web-1 -verbose` | parsing **stops** at the first non-flag, so `-verbose` lands in `fs.Args()` |

### When parsing fails

`flag` prints its own messages to the flag set's output:

- Unknown flag (`-bogus`) or bad value (`-keep abc`): it prints the error
  (`flag provided but not defined: -bogus` or
  `invalid value "abc" for flag -keep: parse error`), then the usage text,
  and `Parse` returns an error.
- `-h`, `-help` or `--help`: it prints the usage text and returns the
  special error `flag.ErrHelp`.

The usage text starts with `Usage of NAME:` and lists every flag. You can
print it yourself with `fs.Usage()`.

### Comparing errors

`errors.Is(err, target)` reports whether `err` is (or wraps) `target`:

```go
if errors.Is(err, os.ErrNotExist) {
	fmt.Println("no such file")
}
```

## Your task

Your team's `deploy` tool has `ParseArgs`, which turns arguments into an
`Options` struct, and `Run`, which uses it. The `-service` flag is already
defined as an example.

### 1. Define the remaining flags

**Goal:** in `ParseArgs`, define three flags and fill in every field of `Options`.

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

**Think it through:**
- Copy the pattern of the `-service` line. Which `fs.` function fits each type?
- Each field needs the value behind the pointer, not the pointer.

### 2. Require -service

**Goal:** after parsing, if the service is empty, print an error plus usage and return an error.

Print `deploy: -service is required` to stderr, then call `fs.Usage()`, then
return an empty `Options{}` and a non-nil error (`errors.New` is fine; add
`"errors"` to the imports). `-service=` (empty) counts as missing.

```text
ParseArgs([]string{"-replicas", "2"}, stderr) // => Options{}, error
stderr:
deploy: -service is required
Usage of deploy:
  ...the flag list...
```

**Think it through:**
- Does this check go before or after `fs.Parse`? Why?
- Print the message only once: `Run` should not print it again.

### 3. Return the right exit code

**Goal:** when `ParseArgs` fails, `Run` returns `0` for `flag.ErrHelp` and `2` for any other error.

Right now it always returns `1`. Print nothing extra: the message is already
on stderr. On success `Run` prints one line to stdout and returns `0`:

```text
Run([]string{"-service", "api", "-replicas", "3", "web-1", "web-2"}, ...) // => 0
stdout: deploy api: replicas=3 timeout=30s dry-run=false hosts=web-1,web-2

Run([]string{"-h"}, ...)     // => 0, usage text on stderr
Run([]string{"-bogus"}, ...) // => 2, error and usage text on stderr
```

A duration prints as `30s`, `5m0s`, `1m30s`. With no hosts, the line ends in
`hosts=`. The success line is already written.

**Think it through:**
- Asking for help isn't a failure. Which error means "help"?
- Which function compares an error with a known one?

## Learn more

- [pkg.go.dev: flag](https://pkg.go.dev/flag)
- [Go by Example: Command-Line Flags](https://gobyexample.com/command-line-flags)
- [pkg.go.dev: errors.Is](https://pkg.go.dev/errors#Is)
- [pkg.go.dev: time.ParseDuration](https://pkg.go.dev/time#ParseDuration)
