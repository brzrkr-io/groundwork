# Subcommands: dispatching git-style commands

`git commit`, `kubectl get`, `docker run`: big ops tools group features into
*subcommands*, each with its own flags. Here you build the skeleton of
`shipctl`, a deploy tool that works the same way.

**New here:** dispatching with `switch`, one FlagSet per subcommand, `fs.NArg` and `fs.Arg`.

## Learn

### What a subcommand is

The first word after the program name picks *which job* to do. Everything
after it belongs to that job:

```text
shipctl deploy -version v1.4.2 -env production api
        ^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        command    the deploy command's own args
```

`Run` receives `os.Args[1:]`, so `args[0]` is the command name and
`args[1:]` is the rest.

### Dispatching with a switch

*Dispatching* means looking at the command name and calling its handler. For
a made-up `cachectl get KEY` / `cachectl put KEY`:

```go
switch args[0] {
case "get", "fetch":
	return runGet(args[1:], stdout, stderr)
case "put":
	return runPut(args[1:], stdout, stderr)
default:
	fmt.Fprintf(stderr, "cachectl: no such command %q\n", args[0])
	return 2
}
```

- Check that `args` isn't empty **before** the switch: `args[0]` on an empty
  slice panics.
- Pass `args[1:]`, so the handler sees only its own flags and arguments.
- One `case` may list several values, separated by commas.

> **Why is it written like this?** `case` values and the `default` keyword
> are Go syntax. Go adds no fall-through, so each case ends by itself. Naming
> handlers `runX` and passing `stdout, stderr` along is the author's choice.

### One FlagSet per subcommand

Each handler builds its own `flag.FlagSet`, as in *Flags: parsing options
with a FlagSet*. That keeps flags apart: `rollback -version v1 api` is an
error, not a silent mix-up.

After `fs.Parse(args)`, the leftover non-flag words are *positional
arguments*. `fs.NArg()` counts them, `fs.Arg(0)` is the first:

```go
fs.Parse([]string{"-ttl", "60", "session-42"})
fs.NArg() // 1
fs.Arg(0) // "session-42"
```

Flags go **before** the service name, because `flag` stops at the first
non-flag.

`runStatus` in `exercise.go` is a finished handler. Read it first: your two
handlers follow the same steps.

### Help is not an error

| Command line | Usage text goes to | Exit code |
|---|---|---|
| `shipctl help`, `-h`, `--help` | stdout | `0` |
| `shipctl` (nothing) | stderr | `2` |
| `shipctl restart api` (unknown) | stderr, after an error line | `2` |

Asked-for help *is* the result, so it goes to stdout. On a wrong command
line, it's a diagnostic, so stderr. The `usage` constant holds the text and
already ends with a newline: print it with `fmt.Fprint(w, usage)`.

## Your task

For now each command only prints what it *would* do. There's no backend yet.

### 1. Dispatch the commands

**Goal:** make `Run` pick the right action from `args[0]`.

| `args` | Do | Return |
|---|---|---|
| empty | print `usage` to stderr | `2` |
| `help`, `-h`, `--help` | print `usage` to stdout | `0` |
| `status`, `deploy`, `rollback` | call `runStatus`, `runDeploy` or `runRollback` with `args[1:]` | what it returns |
| anything else | print `shipctl: unknown command "NAME"`, then `usage`, to stderr | `2` |

```text
$ shipctl restart api
shipctl: unknown command "restart"
usage: shipctl <command> [flags] SERVICE
...
```

The routing test only compares `Run` with the handler, so it passes before
tasks 2 and 3 are done.

**Think it through:**
- What must you check before reading `args[0]`?
- Which verb puts the double quotes around the name?

### 2. Write the deploy command

**Goal:** write `runDeploy`, which prints `deploying SERVICE VERSION to ENV`.

Use a FlagSet named `"deploy"` (with `flag.ContinueOnError` and
`fs.SetOutput(stderr)`, like `runStatus`) and two string flags:

| Flag | Default | Meaning |
|---|---|---|
| `-version` | `""` | version to deploy (required) |
| `-env` | `"staging"` | environment |

Check in this order:

| Problem | Print to stderr | Return |
|---|---|---|
| `Parse` fails with `flag.ErrHelp` (`-h`) | nothing more | `0` |
| `Parse` fails another way | nothing more (`flag` already did) | `2` |
| not exactly one SERVICE | `shipctl: deploy needs exactly one SERVICE` | `2` |
| `-version` empty | `shipctl: deploy needs -version` | `2` |

On success print one line to stdout and return `0`:

```text
$ shipctl deploy -version v1.4.2 api
deploying api v1.4.2 to staging
$ shipctl deploy -env production -version v2.0.0 payments
deploying payments v2.0.0 to production
```

**Think it through:**
- Which lines of `runStatus` can you follow step by step?
- SERVICE comes from `fs.Arg(0)`; VERSION and ENV come from the flag pointers.

### 3. Write the rollback command

**Goal:** write `runRollback`, which prints `rolling back SERVICE by N release(s)`.

Use a FlagSet named `"rollback"` with one int flag, `-steps` (default `1`).
Check in this order:

| Problem | Print to stderr | Return |
|---|---|---|
| `Parse` fails with `flag.ErrHelp` (`-h`) | nothing more | `0` |
| `Parse` fails another way | nothing more | `2` |
| not exactly one SERVICE | `shipctl: rollback needs exactly one SERVICE` | `2` |
| `-steps` below 1 | `shipctl: -steps must be at least 1` | `2` |

```text
$ shipctl rollback api
rolling back api by 1 release(s)
$ shipctl rollback -steps 3 payments
rolling back payments by 3 release(s)
```

Every line ends with a newline (`fmt.Fprintln` adds it for you).

**Think it through:**
- Which `fs.` function makes an int flag?
- `(s)` in `release(s)` is fixed text, always printed.

## Learn more

- [pkg.go.dev: flag.FlagSet](https://pkg.go.dev/flag#FlagSet)
- [Go by Example: Command-Line Subcommands](https://gobyexample.com/command-line-subcommands)
- [Go spec: Switch statements](https://go.dev/ref/spec#Switch_statements)
