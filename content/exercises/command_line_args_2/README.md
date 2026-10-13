# Command-line arguments: fix the port waiter

Deploy scripts often wait for a database or cache to accept connections
before starting an app. A teammate started a `waitport` tool for that, but its
argument handling is buggy. Fix it before it goes into the pipeline.

## Learn

Recap from *Command-line arguments: os.Args and usage errors*: `os.Args[0]` is the program as typed, and
the real arguments start at index 1. Check `len(args)` before indexing,
convert numbers with `strconv.Atoi`, and send usage lines and errors to stderr
with exit code `2`. Results go to stdout with exit code `0`.

New here is an *optional argument*: one the user may leave out. Square
brackets in a usage line mark it, as in `[TIMEOUT_SECONDS]`. So the valid
lengths of `args` are 3 or 4, and when the argument is missing you use a
*default* value:

```go
timeout := 30 // the default
if len(args) == 4 {
	// the user gave one: convert and check it instead
}
```

Another mistake to avoid is ignoring the error from a conversion:
`n, _ := strconv.Atoi(s)` turns `"soon"` into `0` without a word. Always check
`err`.

## Your task

`waitport HOST PORT [TIMEOUT_SECONDS]` checks its arguments and prints what it
would do. It doesn't connect to anything yet. `Run` in `exercise.go` gets the
full `os.Args` and returns the exit code. Every line it prints ends with a
newline. Each task fixes one bug.

### 1. Fix the off-by-one

`Run` reads HOST from `args[0]`, which is the program name. Shift every index
so HOST, PORT and TIMEOUT_SECONDS come from the right place. On success it
prints to stdout and returns `0`:

```text
Run([]string{"waitport", "db.internal", "5432", "10"}, ...) // => 0
stdout: waiting up to 10s for db.internal:5432
```

The port check already prints the right error once it reads the right
argument. A port must be a whole number from 1 to 65535:

```text
Run([]string{"waitport", "db", "0", "5"}, ...)              // => 2
stderr: waitport: port must be 1-65535, got "0"
```

### 2. Send usage errors to stderr with exit code 2

With the wrong number of arguments (anything except 3 or 4), `Run` prints the
usage line to stdout and returns `1`. Print it to **stderr** and return `2`
instead. Nothing goes to stdout. As before, the program name is `args[0]`
exactly as given:

```text
Run([]string{"waitport", "db"}, ...)    // => 2
stderr: usage: waitport HOST PORT [TIMEOUT_SECONDS]

Run([]string{"./waitport"}, ...)        // => 2
stderr: usage: ./waitport HOST PORT [TIMEOUT_SECONDS]
```

### 3. Default and check the timeout

When TIMEOUT_SECONDS is left out, the timeout must be `30`. When it's given,
it must be a whole number of `1` or more. Otherwise print this to stderr and
return `2`:

```text
Run([]string{"waitport", "cache", "6379"}, ...)       // => 0
stdout: waiting up to 30s for cache:6379

Run([]string{"waitport", "db", "5432", "soon"}, ...)  // => 2
stderr: waitport: timeout must be a whole number >= 1, got "soon"

Run([]string{"waitport", "db", "5432", "0"}, ...)     // => 2
stderr: waitport: timeout must be a whole number >= 1, got "0"
```

## Learn more

- [pkg.go.dev: os.Args](https://pkg.go.dev/os#pkg-variables)
- [pkg.go.dev: strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [Go by Example: Command-Line Arguments](https://gobyexample.com/command-line-arguments)
