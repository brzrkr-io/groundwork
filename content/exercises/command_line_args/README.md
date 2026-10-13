# Command-line arguments: os.Args and usage errors

Almost every ops tool starts the same way: someone types `scale api 3` in a
terminal or a CI script. Here you read those words in Go, check them, and give
a clear error when they're wrong.

**New here:** `os.Args`, `strconv.Atoi`, exit codes, printing to stderr.

## Learn

### os.Args

When your program starts, Go puts the words from the command line in
`os.Args`, a `[]string`. For `./scale api 3`:

| Index | Value | What it is |
|---|---|---|
| `os.Args[0]` | `"./scale"` | the program, exactly as typed |
| `os.Args[1]` | `"api"` | the first argument |
| `os.Args[2]` | `"3"` | the second argument |

- The real arguments start at index 1.
- Every argument is a **string**, even `"3"`. You convert numbers yourself.
- The shell splits on spaces. Quotes keep words together:
  `scale "my api" 3` gives `"my api"` as one argument.

Arguments known by their position are called *positional arguments*.

### Check the length before you index

Indexing past the end of a slice crashes the program with a *panic*:

```go
words := []string{"ping"}
fmt.Println(words[1]) // panic: index out of range [1] with length 1
```

So check `len(...)` first, then index.

### Turning "3" into 3

`strconv.Atoi` ("ASCII to integer") turns a string into an `int`. It returns
two values: the number and an error.

```go
n, err := strconv.Atoi("42")  // 42, nil
n, err = strconv.Atoi("abc")  // 0, error
n, err = strconv.Atoi("2.5")  // 0, error
n, err = strconv.Atoi("-7")   // -7, nil: a number, but maybe not a valid one
```

> **Why is it written like this?** Go requires you to receive both return
> values (or skip one with `_`). Calling the error `err` is convention. Using
> `:=` the first time and `=` after is Go's rule: `:=` creates new variables.

### Exit codes and stderr

When the arguments are wrong, a good tool prints a *usage line* (how to call
it) and exits with a non-zero *exit code*. Scripts read that code:

| Exit code | Meaning |
|---|---|
| `0` | success |
| `1` | something failed while running |
| `2` | you called the tool wrong (usage error) |

Results go to *stdout*. Errors and usage lines go to *stderr*, so they never
mix with output another program reads.

`fmt.Fprintf(w, format, values...)` prints to any `io.Writer` `w`:

```go
fmt.Fprintf(stdout, "hello %s, you are %d\n", name, age)
fmt.Fprintf(stderr, "bad name %q\n", name) // %q adds double quotes: "bob"
```

> **Why is it written like this?** The first argument (where to print) is
> required by `Fprintf`. The verbs are fixed by Go: `%s` string, `%d` whole
> number, `%q` quoted string. Text outside the verbs is printed as-is, spaces
> included. `Fprintf` adds no newline, so you write `\n` yourself.

### A testable main

Tests can't easily check a program that reads the real `os.Args` and calls
`os.Exit`. So the logic lives in `Run`, and `main` is one line:

```go
func main() {
	os.Exit(Run(os.Args, os.Stdout, os.Stderr))
}
```

The tests call `Run` with their own arguments and buffers, then check what
was printed and which exit code came back.

## Your task

Your team scales services with `scale SERVICE REPLICAS`. `Run` in
`exercise.go` gets the full `os.Args` (including `args[0]`) and returns the
exit code. Every printed line ends with a newline.

### 1. Print a usage line for the wrong number of arguments

**Goal:** when `len(args)` is not 3, print a usage line to stderr and return `2`.

The line is `usage: PROG SERVICE REPLICAS`, where PROG is `args[0]` exactly as
given. `SERVICE REPLICAS` is fixed text. Print nothing to stdout.

```text
Run([]string{"scale"}, ...)               // => 2
stderr: usage: scale SERVICE REPLICAS

Run([]string{"./scale", "api"}, ...)      // => 2
stderr: usage: ./scale SERVICE REPLICAS
```

Too many arguments (`scale api 3 now`) is also wrong.

**Think it through:**
- Why exactly 3? Count the program name too.
- Which part of the line changes, and which value holds it?

### 2. Reject bad replica counts

**Goal:** if REPLICAS is not a whole number `>= 0`, print an error to stderr and return `2`.

REPLICAS is `args[2]`. Convert it with `strconv.Atoi` (add `"strconv"` to the
imports). It's bad if the conversion fails **or** the number is negative. The
line is `PROG: replicas must be a whole number >= 0, got "VALUE"`, where PROG
is `args[0]` and VALUE is the original text, in double quotes. Print nothing
to stdout.

```text
Run([]string{"scale", "api", "abc"}, ...)  // => 2
stderr: scale: replicas must be a whole number >= 0, got "abc"

Run([]string{"./scale", "api", "-1"}, ...) // => 2
stderr: ./scale: replicas must be a whole number >= 0, got "-1"
```

`"2.5"` and `""` are bad too. `0` is fine.

**Think it through:**
- Which two things can go wrong? Check both in one `if`.
- Which verb adds the double quotes for you?

### 3. Print what you are scaling

**Goal:** with good arguments, print `scaling SERVICE to REPLICAS replicas` to stdout and return `0`.

SERVICE is `args[1]`, REPLICAS is the number you converted. The rest is fixed
text. Print nothing to stderr.

```text
Run([]string{"scale", "api", "3"}, ...)   // => 0
stdout: scaling api to 3 replicas

Run([]string{"scale", "cron", "0"}, ...)  // => 0
stdout: scaling cron to 0 replicas
```

**Think it through:**
- Which two values go into the line, and which verb fits each?
- Did you remove the `not implemented yet` line and `return 1`?

## Learn more

- [pkg.go.dev: os.Args](https://pkg.go.dev/os#pkg-variables)
- [pkg.go.dev: strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [Go by Example: Command-Line Arguments](https://gobyexample.com/command-line-arguments)
- [Go by Example: Exit](https://gobyexample.com/exit)
