## Task 1: guard, then switch

Check `len(args) == 0` first and return early. Then `switch args[0]` with one
`case` per command, a `case "help", "-h", "--help":`, and a `default:` for
unknown commands. Each command case is a single `return runX(args[1:], stdout, stderr)`.

## Tasks 2 and 3: copy runStatus

`runStatus` already has every step you need: create the FlagSet, set its
output, declare flags, parse, check `fs.NArg()`, print. Copy it, rename it,
and change the flags and messages. `fs.Int` works like `fs.String` but gives
you a `*int`.

## Almost the answer for task 2

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

Task 3 is the same shape, with `*steps < 1` as the extra check.
