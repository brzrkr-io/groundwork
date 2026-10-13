# Practice: lay out a tool

`diskcheck` started as one quick `main` function. It works, but nothing in it
can be tested, because everything talks straight to `os.Args`, `os.Stdout`
and `os.Exit`. You'll split it into the layout most Go tools use, and make it
testable on the way.

**New here:** the `cmd/` + `internal/` layout, a thin `main` with `run`,
`io.Writer`, `fmt.Fprintf`.

## Learn

### Recap: module, package, folder

| Word | What it is | Where it's set |
|---|---|---|
| *module* | the whole project, versioned as one | `go.mod` (`module gopherlings.local/project_layout`) |
| *package* | one folder of `.go` files | the `package` line at the top of each file |
| *import path* | module path + folder | `gopherlings.local/project_layout/internal/disk` |

- **Go requires:** every `.go` file in one folder has the same `package`
  name. One folder = one package.
- **Go requires:** a name starting with a capital letter (`Parse`,
  `Reading`) is *exported*: other packages can use it. Lower-case names stay
  inside their package.
- **Go requires:** a package under `internal/` can only be imported by code
  in the same module. The compiler rejects anything else.
- **Convention:** the package name matches the folder name (`internal/disk`
  holds `package disk`), and callers write `disk.Parse`.

### The standard layout

```text
.
├── go.mod
├── cmd/
│   └── backupctl/
│       └── main.go        package main: one folder per program
└── internal/
    └── backup/
        └── backup.go      package backup: the real logic
```

- Every program gets its own folder under `cmd/`, named after the binary.
  `go build ./cmd/backupctl` produces a binary called `backupctl`, and a
  repo can hold several tools side by side.
- The logic lives in packages under `internal/`. It can be tested on its
  own, and every tool in `cmd/` can share it without it becoming a public
  API.

> **Why is it written like this?**
> - **Go requires:** a program is a `package main` with a `func main()`.
>   Since one folder holds one package, each program needs its own folder.
> - **Convention:** the `cmd/<name>/main.go` and `internal/` folders. Go
>   would build other layouts too, but Go developers expect this one.
> - **Author's choice:** the names of the packages under `internal/`.

### A thin main

`main` can't be tested: it has no inputs, no return value, and `os.Exit`
would end the test run. So `main` only connects the real world to a function
that *can* be tested:

```go
func main() {
	os.Exit(serve(os.Args[1:], os.Stdout))
}

func serve(args []string, out io.Writer) int {
	// all the work, writing to out, returning an exit code
}
```

| Piece | Meaning |
|---|---|
| `os.Args` | the command line as a `[]string`. `os.Args[0]` is the program name, so `os.Args[1:]` holds the arguments |
| `io.Writer` | anything you can write to: `os.Stdout`, a file, a `strings.Builder` |
| `fmt.Fprintf(w, …)` | like `Printf`, but writes to `w` |
| `os.Exit(code)` | ends the program with an *exit code*: `0` means success, anything else means failure |

Normal output goes to *stdout*. Errors and usage messages go to *stderr*, so
they don't mix with output another program might read. Tests pass
`strings.Builder`s in as `stdout` and `stderr`, then check what was written.
*Standard streams: stdin, stdout, stderr and exit codes* covers this in
depth.

> **Why is it written like this?** Taking `io.Writer` parameters instead of
> using `os.Stdout` directly is the author's **choice**, and a common Go
> pattern: the function writes wherever its caller says, so a test can catch
> the output.

### Returning errors instead of exiting

A package under `internal/` should never print or exit. It returns an
`error`, and lets the caller decide what to do:

```go
func ParsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("bad port %q", s)
	}
	return n, nil
}
```

`%q` prints a string in double quotes. In a format string, `%%` prints a
single `%`.

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

```text
.
├── go.mod                       module gopherlings.local/project_layout
├── cmd/
│   └── diskcheck/
│       ├── main.go              package main: old all-in-one main + a started run
│       └── main_test.go         tests for run (tasks 3 and 4)
└── internal/
    └── disk/
        ├── disk.go              package disk: Reading, empty Status and Parse
        └── disk_test.go         tests for Status and Parse (tasks 1 and 2)
```

The flow when it's done: `main` → `run` → `disk.Parse` for every argument,
then `disk.Status` for every reading.

Work through the tasks top to bottom. Each one moves a piece of the old
`main` to its new home. The old `main` shows you the logic, so copy from it
freely.

### 1. Move the thresholds into disk.Status

**Goal:** in `internal/disk/disk.go`, finish `Status(percent)` so it returns
`"ok"`, `"warn"` or `"full"`.

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

**Think it through:**
- Find the `status := "ok"` part of the old `main`. Which threshold must you
  check first, so 95 isn't reported as `warn`?

### 2. Move the parsing into disk.Parse

**Goal:** finish `Parse(arg)` so it turns `"web-1=72"` into a `Reading`, and
returns an error for bad input.

```go
disk.Parse("web-1=72")        // => Reading{Host: "web-1", Percent: 72}, nil
disk.Parse("db-1.internal=0") // => Reading{Host: "db-1.internal", Percent: 0}, nil
```

It returns an error, with exactly this text, when:

| Problem | Examples | Error text |
|---|---|---|
| no `=`, or nothing before it | `web-1`, `=72` | `bad reading "web-1": want HOST=PERCENT` |
| percent isn't a whole number from 0 to 100 | `web-1=lots`, `web-1=`, `web-1=120`, `web-1=-5` | `bad percent in "web-1=lots": want 0-100` |

- The quoted part is the whole argument, as `%q` prints it.
- No `diskcheck: ` prefix and no newline: `run` adds those in task 4.
- Add the imports you need (`fmt`, `strconv`, `strings`).

**Think it through:**
- The old `main` already has these checks. What replaces each
  "print, then `os.Exit(2)`" pair, now that `Parse` returns an error?
- What do you return as the `Reading` when there's an error?

### 3. Write run for good input

**Goal:** in `cmd/diskcheck/main.go`, finish `run(args, stdout, stderr)`
so it prints one report line per argument and returns the exit code.

`run` already handles the no-arguments case. Its loop only echoes each
argument for now.

- Import `gopherlings.local/project_layout/internal/disk`.
- Parse every argument with `disk.Parse`, and get each status from
  `disk.Status`.
- Write one line per reading to `stdout`, in order: the host, a space, the
  percent followed by `%`, a space, the status, and a newline.
- Return `1` if any disk is `full`, otherwise `0`.

```go
run([]string{"web-1=72", "web-2=91"}, stdout, stderr)
// stdout: "web-1 72% ok\nweb-2 91% full\n"
// => 1
```

Then make `main` thin: its whole body becomes one line that calls `run` with
the real arguments (without the program name), `os.Stdout` and `os.Stderr`,
and passes the result to `os.Exit`. Delete the imports it no longer needs.
Nothing tests `main`; keeping it this small means nothing in it can break.

**Think it through:**
- Write to the `stdout` parameter, not `os.Stdout`. Which one can the test
  see?
- Task 4 needs every argument parsed before anything is printed. Can you
  collect the readings first, then print them in a second loop?

### 4. Handle bad input in run

**Goal:** when `disk.Parse` fails, `run` writes the error to `stderr`,
prints no report, and returns `2`.

| Situation | stderr | stdout | Returns |
|---|---|---|---|
| no arguments (already done) | `usage: diskcheck HOST=PERCENT...` + newline | nothing | `2` |
| `disk.Parse` fails | `diskcheck: ` + the error text + newline | nothing | `2` |

A bad *last* argument still means no report at all:

```go
run([]string{"web-1=72", "web-2"}, stdout, stderr)
// stdout: ""
// stderr: "diskcheck: bad reading \"web-2\": want HOST=PERCENT\n"
// => 2
```

**Think it through:**
- The no-arguments check at the top of `run` already shows the shape: write
  to `stderr`, then return `2`.
- Which parts of the stderr line are fixed text, and which part is the
  error?

## Learn more

- [Organizing a Go module](https://go.dev/doc/modules/layout)
- [Package io: Writer](https://pkg.go.dev/io#Writer)
- [Go by Example: Exit](https://gobyexample.com/exit)
