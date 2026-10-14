## Task 1: Dispatch the commands

Check for an empty `args` first, because `args[0]` would panic. Then pick the
handler based on `args[0]`, and give it only what comes after the name.

### Closer

Return early on `len(args) == 0`, then `switch args[0]` with
`case "help", "-h", "--help":`, one case per command
(`return runStatus(args[1:], stdout, stderr)`), and a `default:` for unknown
names. Print the help with `fmt.Fprint(w, usage)`.

### Answer

```go
if len(args) == 0 {
	fmt.Fprint(stderr, usage)
	return 2
}
switch args[0] {
case "help", "-h", "--help":
	fmt.Fprint(stdout, usage)
	return 0
case "status":
	return runStatus(args[1:], stdout, stderr)
case "deploy":
	return runDeploy(args[1:], stdout, stderr)
case "rollback":
	return runRollback(args[1:], stdout, stderr)
default:
	fmt.Fprintf(stderr, "shipctl: unknown command %q\n", args[0])
	fmt.Fprint(stderr, usage)
	return 2
}
```

## Task 2: Write the deploy command

`runStatus` already has every step: create the FlagSet, set its output, define
flags, parse, check `fs.NArg()`, print. Copy it and change the flags and
messages, then add the `-version` check after the SERVICE check.

### Closer

Define `version := fs.String("version", "", ...)` and
`env := fs.String("env", "staging", ...)`. After the `fs.NArg() != 1` check,
add `if *version == "" { ... }`.

### Answer

```go
fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
fs.SetOutput(stderr)
version := fs.String("version", "", "version to deploy (required)")
env := fs.String("env", "staging", "environment to deploy to")
if err := fs.Parse(args); err != nil {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 2
}
if fs.NArg() != 1 {
	fmt.Fprintln(stderr, "shipctl: deploy needs exactly one SERVICE")
	return 2
}
if *version == "" {
	fmt.Fprintln(stderr, "shipctl: deploy needs -version")
	return 2
}
fmt.Fprintf(stdout, "deploying %s %s to %s\n", fs.Arg(0), *version, *env)
return 0
```

## Task 3: Write the rollback command

This is the same shape as `runDeploy`, with one int flag instead of two string
flags, and a range check instead of the `-version` check.

### Closer

`steps := fs.Int("steps", 1, ...)` gives you a `*int`. After the SERVICE check,
reject `*steps < 1`. Print the count with `%d`.

### Answer

```go
fs := flag.NewFlagSet("rollback", flag.ContinueOnError)
fs.SetOutput(stderr)
steps := fs.Int("steps", 1, "how many releases to roll back")
if err := fs.Parse(args); err != nil {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 2
}
if fs.NArg() != 1 {
	fmt.Fprintln(stderr, "shipctl: rollback needs exactly one SERVICE")
	return 2
}
if *steps < 1 {
	fmt.Fprintln(stderr, "shipctl: -steps must be at least 1")
	return 2
}
fmt.Fprintf(stdout, "rolling back %s by %d release(s)\n", fs.Arg(0), *steps)
return 0
```
