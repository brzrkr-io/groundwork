# Practice: lay out a tool

`diskcheck` started as one quick `main` function. It works, but nothing in it
can be tested, because everything talks straight to `os.Args`, `os.Stdout`
and `os.Exit`. You'll split it into the layout most Go tools use, and make it
testable on the way.

## Learn

### Recap

A package is a folder, and its import path is the module path plus the
folder. Capitalised names are exported. Packages under `internal/` can only
be imported from inside the module.

### The standard layout

```text
.
├── go.mod
├── cmd/
│   └── diskcheck/
│       └── main.go        package main: one folder per program
└── internal/
    └── disk/
        └── disk.go        package disk: the real logic
```

- Every program gets its own folder under `cmd/`, named after the binary.
  `go build ./cmd/diskcheck` produces a binary called `diskcheck`, and a
  repo can hold several tools side by side.
- Logic lives in packages under `internal/`. It can be tested on its own, and
  shared by all the tools in `cmd/` without becoming a public API.

### A thin main

`main` can't be tested: it has no inputs and no return value, and `os.Exit`
would end the test run. So `main` does nothing but connect the real world to
a function that *can* be tested:

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	// all the work, writing to stdout and stderr, returning an exit code
}
```

New pieces:

| Piece | Meaning |
|---|---|
| `os.Args` | the command line as a `[]string`. `os.Args[0]` is the program name, so `os.Args[1:]` holds the arguments |
| `io.Writer` | anything you can write bytes to: `os.Stdout`, a file, a `strings.Builder` |
| `fmt.Fprintf(w, …)` | like `Printf`, but writes to `w` |
| `os.Exit(code)` | ends the program with an *exit code*: `0` means success, anything else means failure |

Normal output goes to *stdout*. Errors and usage messages go to *stderr*, so
they don't mix with output another program might read. Tests pass
`strings.Builder`s as `stdout` and `stderr`, then check what was written.
*CLI tools and the OS* covers all of this in depth.

## Your task

`diskcheck` takes one `HOST=PERCENT` argument per disk and reports on each:

```bash
$ go run ./cmd/diskcheck web-1=72 db-1=85 web-2=91
web-1 72% ok
db-1 85% warn
web-2 91% full
$ echo $?
1
```

### File map

| File | What's in it |
|---|---|
| `cmd/diskcheck/main.go` | the old all-in-one `main`, plus a started `run` |
| `cmd/diskcheck/main_test.go` | tests for `run` (tasks 3 and 4) |
| `internal/disk/disk.go` | `Reading`, plus empty `Status` and `Parse` |
| `internal/disk/disk_test.go` | tests for `Status` and `Parse` (tasks 1 and 2) |

Suggested order: work through the tasks top to bottom. Each one moves a
piece of the old `main` to its new home. The old `main` shows you the logic,
so copy from it freely.

### 1. Move the thresholds into disk.Status

In `internal/disk/disk.go`, finish `Status(percent int) string`:

| Percent | Status |
|---|---|
| below 80 | `ok` |
| 80 to 89 | `warn` |
| 90 or more | `full` |

```go
disk.Status(72) // => "ok"
disk.Status(80) // => "warn"
disk.Status(90) // => "full"
```

### 2. Move the parsing into disk.Parse

Finish `Parse(arg string) (Reading, error)`:

```go
disk.Parse("web-1=72") // => Reading{Host: "web-1", Percent: 72}, nil
```

It returns an error, with exactly this text, when:

| Problem | Example | Error text |
|---|---|---|
| no `=`, or nothing before it | `web-1`, `=72` | `bad reading "web-1": want HOST=PERCENT` |
| percent isn't a whole number from 0 to 100 | `web-1=lots`, `web-1=120` | `bad percent in "web-1=lots": want 0-100` |

The quoted part is the whole argument, as `%q` prints it. Use `fmt.Errorf`
and add the imports you need (`fmt`, `strconv`, `strings`).

### 3. Write run for good input

In `cmd/diskcheck/main.go`, finish `run(args, stdout, stderr)` for input
without mistakes. Import `gopherlings.local/project_layout/internal/disk`.

`run` already handles the no-arguments case: it writes the usage line to
`stderr` and returns `2`, where the old `main` called `os.Exit(2)`. Its loop
only echoes each argument for now.

- Parse every argument with `disk.Parse`.
- Write one line per reading to `stdout`, in order:
  `<host> <percent>% <status>`, followed by a newline. Use `disk.Status`.
- Return `1` if any disk is `full`, otherwise `0`.

```go
run([]string{"web-1=72", "web-2=91"}, stdout, stderr)
// stdout: "web-1 72% ok\nweb-2 91% full\n"
// => 1
```

In a format string, `%%` prints a single `%`.

Then make `main` thin: replace its whole body with
`os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`, and delete the imports
it no longer needs. Nothing tests `main`. Keeping it this small means there's
nothing in it that could break.

### 4. Handle bad input in run

| Situation | stderr | stdout | Returns |
|---|---|---|---|
| no arguments (already done) | `usage: diskcheck HOST=PERCENT...` + newline | nothing | `2` |
| `disk.Parse` fails | `diskcheck: ` + the error text + newline | nothing | `2` |

Parse *all* arguments before you print anything, so a bad last argument
means no report at all:

```go
run([]string{"web-1=72", "web-2"}, stdout, stderr)
// stdout: ""
// stderr: "diskcheck: bad reading \"web-2\": want HOST=PERCENT\n"
// => 2
```

## Learn more

- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Package io: Writer](https://pkg.go.dev/io#Writer)
- [Go by Example: Exit](https://gobyexample.com/exit)
