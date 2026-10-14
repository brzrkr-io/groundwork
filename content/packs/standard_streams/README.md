# Standard streams: stdin, stdout, stderr and exit codes

Unix tools are small programs glued together with pipes. A good pipeline
citizen reads stdin, puts results on stdout, puts problems on stderr, and
reports success or failure with its exit code.

**New here:** stdin as an `io.Reader`. Stream separation and exit code `1` vs `2` are recaps from *Command-line arguments: os.Args and usage errors*, now applied to a tool in a pipeline.

## Learn

### Three streams

| Stream | Go in `main` | Used for |
|---|---|---|
| *stdin* | `os.Stdin` | data coming in: a pipe, a file with `<`, the keyboard |
| *stdout* | `os.Stdout` | **results**: what the next program reads |
| *stderr* | `os.Stderr` | **diagnostics**: errors, warnings, usage text |

Pipes only connect stdout:

```bash
cat hosts.txt | hostfilter web- | xargs -n1 ssh-keyscan
```

A warning on stdout would be fed to `ssh-keyscan` as a host. On stderr it
still shows in your terminal, and `2>` can redirect it:

```bash
hostfilter web- < hosts.txt > web-hosts.txt 2> errors.log
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | runtime failure: bad data, unreadable input |
| `2` | usage error: wrong arguments |

Shells act on them: `cmd && echo ok` runs `echo` only on success,
`cmd || echo failed` only on failure, and `set -e` stops a script at the
first failure. A wrong exit code is a real bug.

### Wiring it up

As before, the logic lives in `Run`, which gets the streams as parameters:

```go
func main() {
	os.Exit(streams.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

`os.Stdin` is an `io.Reader`; `os.Stdout` and `os.Stderr` are `io.Writer`s.
Tests pass fakes: `strings.NewReader("web-01\n")` for stdin and
`strings.Builder`s for the outputs.

> **Why is it written like this?** Passing streams as parameters is the
> author's choice so tests can swap them. Go requires you to write to the
> parameter (`stdout`), not `os.Stdout`, or the test never sees the output.

### Reading stdin

Read line by line with `bufio.Scanner`, as in *Reading input line by line*:
count lines at the top of the loop, `strings.TrimSpace` each one, and check
`sc.Err()` after the loop.

Two `strings` helpers you'll want:

```go
strings.HasPrefix("api-01", "api") // true
strings.ContainsAny("a b", " \t")  // true: has a space or a tab
```

To test read errors, the tests glue `iotest.ErrReader` (a reader that always
fails) after normal text with `io.MultiReader`: a stdin that gives a few
lines, then breaks.

## Your task

`hostfilter PREFIX` reads hostnames from stdin, one per line, and prints the
ones that start with PREFIX. Add `"bufio"` and `"strings"` to the imports
when you need them.

### 1. Fix the usage check

**Goal:** with anything other than exactly one argument, print `usage: hostfilter PREFIX` to stderr and return `2`.

Right now `Run` prints the usage line to stdout, returns `0`, and lets
`hostfilter web- db-` through. Print nothing to stdout.

```text
Run([]string{"web-", "db-"}, ...) // => 2
stderr: usage: hostfilter PREFIX
```

**Think it through:**
- The check catches 0 arguments. What about 2 or 3?
- Which `Fprintln` argument picks the stream?

### 2. Filter the hosts

**Goal:** in `Filter`, print each stdin host that starts with `prefix` to stdout, one per line, and return `0`.

For each line:

1. Trim spaces and tabs from both ends.
2. Skip it if it's blank or starts with `#`.
3. If it starts with `prefix`, print it on its own line.

```text
$ printf '# web tier\n  web-01  \ndb-01\n\nweb-02\n' | hostfilter web-
web-01
web-02
```

The last line may have no newline; it still counts.

**Think it through:**
- Which parameter is the prefix, and which is the input?
- Print the trimmed host, not the raw line.

### 3. Report problems on stderr

**Goal:** report invalid hosts and read errors on stderr, and return `1` if there were any.

A host that still contains a space or tab after trimming is invalid. Check it
after skipping blanks and comments, before the prefix check. For each one,
print to stderr:

```text
hostfilter: line 2: invalid host "web 02"
```

- The line number counts **every** line from 1, blanks and comments too.
- The host uses `%q`, so a tab shows as `\t`.
- Don't stop: keep printing good hosts, then return `1` at the end.

If `sc.Err()` returns an error after the loop, print this to stderr and
return `1` (the part after `reading stdin: ` is the error, via `%v`):

```text
hostfilter: reading stdin: disk read failed
```

Hosts printed before the error stay printed.

**Think it through:**
- How do you remember "something was bad" until the loop ends?
- Where must the line counter go up so skipped lines still count?

## Learn more

- [pkg.go.dev: os.Stdin, Stdout, Stderr](https://pkg.go.dev/os#pkg-variables)
- [pkg.go.dev: testing/iotest.ErrReader](https://pkg.go.dev/testing/iotest#ErrReader)
- [Go by Example: Line Filters](https://gobyexample.com/line-filters)
- [Go by Example: Exit](https://gobyexample.com/exit)
