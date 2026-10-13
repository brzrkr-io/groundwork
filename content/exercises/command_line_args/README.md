# Command-line arguments: os.Args and usage errors

Almost every ops tool starts the same way: someone types
`scale api 3` in a terminal or a CI script. In this exercise you read those
words in Go, check them, and give a clear error when they're wrong.

## Learn

### os.Args

When your program starts, Go puts the words from the command line in
`os.Args`, a `[]string`. For this command:

```text
$ ./scale api 3
```

`os.Args` is:

| Index | Value | What it is |
|---|---|---|
| `os.Args[0]` | `"./scale"` | the program, exactly as it was typed |
| `os.Args[1]` | `"api"` | the first argument |
| `os.Args[2]` | `"3"` | the second argument |

A few things to notice:

- `os.Args[0]` is the program itself, so the real arguments start at index 1.
  It's written exactly as typed: `./scale`, `scale` or `/usr/local/bin/scale`.
- Every argument is a **string**, even `"3"`. You convert numbers yourself.
- The shell splits the line on spaces before Go ever sees it. Quotes keep
  words together: `scale "my api" 3` gives `"my api"` as one argument.

Arguments identified by their position (first, second...) are called
*positional arguments*.

### Check the length before you index

Indexing past the end of a slice crashes the program with a *panic*:

```go
args := []string{"scale"}
fmt.Println(args[1]) // panic: index out of range [1] with length 1
```

So always check `len(args)` first. `scale SERVICE REPLICAS` needs exactly
3 entries: the program plus two arguments.

### Turning "3" into 3

`strconv.Atoi` ("ASCII to integer") converts a string to an `int`, and returns
an error when the text is not a whole number:

```go
n, err := strconv.Atoi("3")   // 3, nil
n, err = strconv.Atoi("abc")  // 0, error
n, err = strconv.Atoi("2.5")  // 0, error
n, err = strconv.Atoi("-1")   // -1, nil: a valid number, but maybe not a valid count
```

### Usage messages, stderr and exit codes

When the arguments are wrong, a good tool prints a *usage line* that shows how
to call it, and exits with a non-zero *exit code*. Scripts and CI read that
code to tell success from failure:

| Exit code | Meaning |
|---|---|
| `0` | success |
| `1` | something failed while running |
| `2` | you called the tool wrong (usage error) |

Results go to *standard output* (stdout). Errors and usage lines go to
*standard error* (stderr), so they never get mixed into output that another
program reads. `fmt.Fprintln(w, ...)` and `fmt.Fprintf(w, ...)` print to any
`io.Writer` `w`, such as stdout or stderr. `Fprintln` adds a newline at the
end; with `Fprintf` you add `\n` yourself.

### A testable main

Tests can't easily check a program that reads the real `os.Args` and calls
`os.Exit`. So, as in earlier exercises, all the logic lives in a `Run`
function, and `main` is one line:

```go
func main() {
	os.Exit(Run(os.Args, os.Stdout, os.Stderr))
}
```

The tests call `Run` directly with their own arguments, and pass in buffers
(`bytes.Buffer`) instead of the real stdout and stderr. Afterwards they check
what was written and which exit code came back.

## Your task

Your team scales services with a tiny tool, `scale SERVICE REPLICAS`. `Run` in
`exercise.go` gets the full `os.Args` (including `args[0]`) and returns the
exit code. Every line it prints ends with a newline.

### 1. Print a usage line for the wrong number of arguments

When `len(args)` is not 3, print this line to **stderr**, using `args[0]` for
the program name exactly as given, and return `2`. Print nothing to stdout.

```text
Run([]string{"scale"}, ...)               // => 2
stderr: usage: scale SERVICE REPLICAS

Run([]string{"./scale", "api"}, ...)      // => 2
stderr: usage: ./scale SERVICE REPLICAS
```

### 2. Reject bad replica counts

REPLICAS must be a whole number that is `0` or more. Convert it with
`strconv.Atoi`. If the conversion fails, or the number is negative, print this
line to **stderr** and return `2` (a bad argument is a usage error). The bad
value is shown in double quotes, which `%q` does for you. Print nothing to
stdout.

```text
Run([]string{"scale", "api", "abc"}, ...) // => 2
stderr: scale: replicas must be a whole number >= 0, got "abc"

Run([]string{"./scale", "api", "-1"}, ...) // => 2
stderr: ./scale: replicas must be a whole number >= 0, got "-1"
```

`0` is allowed: `scale worker 0` prints `scaling worker to 0 replicas`.

### 3. Print what you are scaling

When the arguments are good, print this line to **stdout** and return
`0`. Print nothing to stderr.

```text
Run([]string{"scale", "api", "3"}, ...)   // => 0
stdout: scaling api to 3 replicas
```

## Learn more

- [pkg.go.dev: os.Args](https://pkg.go.dev/os#pkg-variables)
- [pkg.go.dev: strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
- [Go by Example: Command-Line Arguments](https://gobyexample.com/command-line-arguments)
- [Go by Example: Exit](https://gobyexample.com/exit)
