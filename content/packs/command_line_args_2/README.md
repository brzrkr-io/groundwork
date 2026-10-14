# Command-line arguments: fix the port waiter

Deploy scripts often wait for a database to accept connections before
starting an app. A teammate started a `waitport` tool for that, but its
argument handling is buggy. Fix it.

**New here:** optional arguments and default values.

## Learn

### Recap

As in *Command-line arguments: os.Args and usage errors*:

- `args[0]` is the program as typed; real arguments start at index 1.
- Check `len(args)` before indexing.
- Convert numbers with `strconv.Atoi` and check `err`.
- Usage lines and errors go to stderr with exit code `2`. Results go to
  stdout with exit code `0`.

### Optional arguments

An *optional argument* is one the user may leave out. Square brackets in a
usage line mark it: `greet NAME [GREETING]`. When it's missing, you use a
*default* value:

```go
greeting := "hello" // the default
if len(args) == 3 {
	greeting = args[2] // the user gave one
}
```

> **Why is it written like this?** Setting the default first, then
> overwriting it, is a common Go pattern (author's choice, but you'll see it
> everywhere). Inside the `if` it's `=`, not `:=`: Go requires `=` to change an
> existing variable. `:=` would make a new one that disappears at `}`.

### Don't ignore errors

`n, _ := strconv.Atoi(s)` throws the error away, so `"soon"` silently
becomes `0`. Always keep `err` and check it.

## Your task

`waitport HOST PORT [TIMEOUT_SECONDS]` checks its arguments and prints what it
would do. `Run` in `exercise.go` gets the full `os.Args` and returns the exit
code. Every printed line ends with a newline. Each task fixes one bug.

### 1. Fix the off-by-one

**Goal:** make `Run` read HOST, PORT and TIMEOUT_SECONDS from the right indexes.

Right now HOST comes from `args[0]`, the program name. Shift every index by
one, including in the port error message. On success it prints to stdout and
returns `0`:

```text
Run([]string{"waitport", "db.internal", "5432", "10"}, ...) // => 0
stdout: waiting up to 10s for db.internal:5432
```

A port must be a whole number from 1 to 65535. The existing check prints
this once it reads the right argument:

```text
Run([]string{"waitport", "db", "0", "5"}, ...)              // => 2
stderr: waitport: port must be 1-65535, got "0"
```

**Think it through:**
- Write out `args` for the example above, index by index. Where is HOST?
- Find every `args[...]` in `Run`. The one in the error message counts too.

### 2. Send usage errors to stderr with exit code 2

**Goal:** with the wrong number of arguments, print the usage line to stderr and return `2`.

Valid lengths are 3 or 4. Right now the usage line goes to stdout and `Run`
returns `1`. Nothing should go to stdout. The program name is `args[0]` as
given:

```text
Run([]string{"waitport", "db"}, ...)    // => 2
stderr: usage: waitport HOST PORT [TIMEOUT_SECONDS]

Run([]string{"./waitport"}, ...)        // => 2
stderr: usage: ./waitport HOST PORT [TIMEOUT_SECONDS]
```

**Think it through:**
- Which argument of `Fprintf` decides where the text goes?
- The message itself is already right; only two things change.

### 3. Default and check the timeout

**Goal:** use `30` when TIMEOUT_SECONDS is missing; otherwise it must be a whole number `>= 1`.

TIMEOUT_SECONDS is the 4th entry of `args`. If it fails to convert or is
below 1, print `PROG: timeout must be a whole number >= 1, got "VALUE"` to
stderr and return `2`. VALUE is the original text in double quotes.

```text
Run([]string{"waitport", "cache", "6379"}, ...)       // => 0
stdout: waiting up to 30s for cache:6379

Run([]string{"waitport", "db", "5432", "soon"}, ...)  // => 2
stderr: waitport: timeout must be a whole number >= 1, got "soon"

Run([]string{"waitport", "db", "5432", "0"}, ...)     // => 2
stderr: waitport: timeout must be a whole number >= 1, got "0"
```

**Think it through:**
- Where is the default set now, and what number is it?
- What does `_` throw away? Look at how the port check handles the same case.

## Learn more

- [pkg.go.dev: os.Args](https://pkg.go.dev/os#pkg-variables)
- [pkg.go.dev: strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [Go by Example: Command-Line Arguments](https://gobyexample.com/command-line-arguments)
