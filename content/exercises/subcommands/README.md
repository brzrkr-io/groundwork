# Subcommands: dispatching git-style commands

`git commit`, `kubectl get`, `docker run`: big ops tools group their features
into *subcommands*, each with its own flags. In this exercise you build the
skeleton of a deploy tool, `shipctl`, that works the same way.

## Learn

### What a subcommand is

A subcommand is the first word after the program name. It picks *which job*
the tool does, and everything after it belongs to that job:

```text
shipctl deploy -version v1.4.2 -env production api
        ^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        command    the deploy command's own args
```

As in *Flags: parsing options with a FlagSet*, `Run` receives `os.Args[1:]`, so `args[0]` is the command name
and `args[1:]` is everything that belongs to it.

### Dispatching with a switch

*Dispatching* means looking at the command name and calling the function that
handles it. A `switch` (from *Switch*) does this neatly:

```go
switch args[0] {
case "status":
	return runStatus(args[1:], stdout, stderr)
case "deploy":
	return runDeploy(args[1:], stdout, stderr)
default:
	fmt.Fprintf(stderr, "shipctl: unknown command %q\n", args[0])
	return 2
}
```

Two details matter:

- Check `len(args) == 0` **before** the switch. `args[0]` on an empty slice
  panics.
- Pass `args[1:]`, not `args`. The handler should only see its own flags and
  arguments, never its own name.

A `case` can list several values: `case "help", "-h", "--help":` matches any
of the three.

### One FlagSet per subcommand

Each handler builds its own `flag.FlagSet`, exactly as in *Flags: parsing options with a FlagSet*. That keeps
the flags apart: `deploy` has `-version` and `rollback` has `-steps`, and
`shipctl rollback -version v1 api` is an error instead of a silent mix-up.

After `fs.Parse(args)`, the arguments that are not flags are the
*positional arguments*. `fs.NArg()` returns how many there are and
`fs.Arg(0)` returns the first one:

```go
fs.Parse([]string{"-env", "production", "api"})
fs.NArg() // 1
fs.Arg(0) // "api"
```

`flag` stops at the first argument that is not a flag, so flags go **before**
the service name: `deploy -version v1 api`, not `deploy api -version v1`.

`runStatus` in `exercise.go` is a finished handler. Read it first: your two
handlers follow the same steps.

### Help is not an error

| Command line | Where the usage text goes | Exit code |
|---|---|---|
| `shipctl help`, `shipctl -h`, `shipctl --help` | stdout | `0` |
| `shipctl` (nothing) | stderr | `2` |
| `shipctl restart api` (unknown) | stderr, after an error line | `2` |

When someone asks for help, the help *is* the result, so it goes to stdout and
the tool succeeds. When the command line is wrong, the usage text is a
diagnostic, so it goes to stderr with exit code `2` (usage error). The
`usage` constant in `exercise.go` holds the text. It already ends with a
newline, so print it with `fmt.Fprint(w, usage)`.

## Your task

Your team wants one tool for day-to-day service operations. For now, each
command only prints what it *would* do. There is no backend yet.

### 1. Dispatch the commands

Write `Run`:

- No arguments: print `usage` to stderr and return `2`.
- `help`, `-h` or `--help`: print `usage` to stdout and return `0`.
- `status`, `deploy`, `rollback`: call `runStatus`, `runDeploy` or
  `runRollback` with `args[1:]` and return what it returns.
- Anything else: print an error line, then `usage`, to stderr and return `2`.
  The error line uses `%q`:

```text
$ shipctl restart api
shipctl: unknown command "restart"
usage: shipctl <command> [flags] SERVICE
...
```

The test also checks that `Run` passes `deploy` and `rollback` command lines
on to their handlers, so it passes before tasks 2 and 3 are done.

### 2. Write the deploy command

Write `runDeploy`. Use a FlagSet named `"deploy"` (with
`flag.ContinueOnError` and `fs.SetOutput(stderr)`, like `runStatus`) and two
string flags:

| Flag | Default | Meaning |
|---|---|---|
| `-version` | `""` | version to deploy (required) |
| `-env` | `"staging"` | environment to deploy to |

Check, in this order:

| Problem | Print to stderr | Return |
|---|---|---|
| `fs.Parse` fails with `flag.ErrHelp` (`-h`) | nothing more | `0` |
| `fs.Parse` fails in any other way | nothing more (`flag` already did) | `2` |
| not exactly one SERVICE | `shipctl: deploy needs exactly one SERVICE` | `2` |
| `-version` empty | `shipctl: deploy needs -version` | `2` |

On success print one line to stdout and return `0`:

```text
$ shipctl deploy -version v1.4.2 api
deploying api v1.4.2 to staging
$ shipctl deploy -env production -version v2.0.0 payments
deploying payments v2.0.0 to production
```

### 3. Write the rollback command

Write `runRollback`. Use a FlagSet named `"rollback"` with one int flag,
`-steps` (default `1`): how many releases to go back. Check, in this order:

| Problem | Print to stderr | Return |
|---|---|---|
| `fs.Parse` fails with `flag.ErrHelp` (`-h`) | nothing more | `0` |
| `fs.Parse` fails in any other way | nothing more | `2` |
| not exactly one SERVICE | `shipctl: rollback needs exactly one SERVICE` | `2` |
| `-steps` below 1 | `shipctl: -steps must be at least 1` | `2` |

On success:

```text
$ shipctl rollback api
rolling back api by 1 release(s)
$ shipctl rollback -steps 3 payments
rolling back payments by 3 release(s)
```

Every error line ends with a newline (`fmt.Fprintln` adds it for you).

## Learn more

- [flag.FlagSet](https://pkg.go.dev/flag#FlagSet)
- [Go by Example: Command-Line Subcommands](https://gobyexample.com/command-line-subcommands)
- [Go spec: Switch statements](https://go.dev/ref/spec#Switch_statements)
