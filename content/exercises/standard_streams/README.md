# Standard streams: stdin, stdout, stderr and exit codes

Unix tools are small programs glued together with pipes. To be a good
citizen in a pipeline, your Go tool must read from stdin, put results on
stdout, put problems on stderr, and report success or failure with its exit
code.

## Learn

### Three streams

Every process starts with three open streams:

| Stream | Go in `main` | Used for |
|---|---|---|
| *stdin* (standard input) | `os.Stdin` | data coming in: a pipe, a file with `<`, or the keyboard |
| *stdout* (standard output) | `os.Stdout` | **results**: the data the next program consumes |
| *stderr* (standard error) | `os.Stderr` | **diagnostics**: errors, warnings, usage text |

Why two output streams? Because pipes only connect stdout:

```bash
cat hosts.txt | hostfilter web- | xargs -n1 ssh-keyscan
```

Only the hosts flow into `xargs`. If `hostfilter` printed a warning to stdout,
`xargs` would try to `ssh-keyscan` the warning text. Warnings on stderr still
show up in your terminal, and you can send them somewhere else with `2>`:

```bash
hostfilter web- < hosts.txt > web-hosts.txt 2> errors.log
```

### Exit codes

When a program ends, it hands the shell a number: its *exit code*. `0` means
success. Anything else means failure. This module uses:

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | runtime failure: bad data, unreadable input |
| `2` | usage error: wrong arguments |

The shell keeps the last exit code in `$?`, and uses it for control flow:

```bash
hostfilter web- < hosts.txt; echo $?   # prints the exit code
hostfilter web- < hosts.txt && echo ok     # runs echo only on success
hostfilter web- < hosts.txt || echo failed # runs echo only on failure
set -e  # in a script: stop at the first command that fails
```

So a wrong exit code is a real bug: a CI job or deploy script will carry on
when it should stop, or stop when it should carry on.

### Wiring it up, and testing it

As in *Flags: parsing options with a FlagSet*, all the logic lives in `Run`, and it receives the streams as
arguments. `main` is one line:

```go
func main() {
	os.Exit(streams.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

`os.Stdin` is an `io.Reader`, and `os.Stdout` and `os.Stderr` are
`io.Writer`s. The tests pass fakes instead: `strings.NewReader("web-01\n")`
as stdin, and two `strings.Builder`s that collect whatever you write, so they
can compare stdout and stderr separately. Write to a stream with the `F`
functions: `fmt.Fprintln(stdout, host)`, `fmt.Fprintf(stderr, ...)`.

### Reading stdin

Read stdin line by line with `bufio.Scanner`, exactly as in *Reading input
line by line*: count lines at the top of the loop, `strings.TrimSpace` each
one, and check `sc.Err()` after the loop.

To test that read errors are handled, the tests use `iotest.ErrReader` from
the `testing/iotest` package. It is a reader whose `Read` always fails with
the error you give it. Glued after a normal reader with `io.MultiReader`, it
acts like a stdin that delivers a few lines and then breaks.

## Your task

`hostfilter PREFIX` reads hostnames from stdin, one per line, and prints the
ones that start with `PREFIX`. Add `"bufio"` and `"strings"` to the imports
when you need them.

### 1. Fix the usage check

`Run` needs exactly one argument. Right now it prints its usage line to
stdout, exits `0`, and lets `hostfilter web- db-` through. With anything other
than exactly one argument, print this line to **stderr** and return `2`:

```text
usage: hostfilter PREFIX
```

### 2. Filter the hosts

Write `Filter`. For each line of stdin:

1. Trim spaces and tabs from both ends.
2. Skip it if it is blank or starts with `#`.
3. If it starts with `prefix`, print it to stdout on its own line.

Return `0`.

```text
$ printf '# web tier\n  web-01  \ndb-01\n\nweb-02\n' | hostfilter web-
web-01
web-02
$ echo $?
0
```

### 3. Report problems on stderr

A host that still contains a space or a tab after trimming is invalid
(`strings.ContainsAny(host, " \t")` checks for both). Check this after
skipping blank and comment lines, and before the prefix check. For each
invalid line, print to stderr:

```text
hostfilter: line 2: invalid host "web 02"
```

The line number counts **every** line from 1, including blank and comment
lines. The host is printed with `%q`, so a tab shows up as `\t`. Do not
stop: keep printing the good hosts, and return `1` at the end if any line was
invalid.

If `sc.Err()` returns an error after the loop, print the following to stderr
and return `1`:

```text
hostfilter: reading stdin: disk read failed
```

The part after `reading stdin: ` is the error itself (`%v`). Hosts printed
before the error stay printed.

## Learn more

- [os package: Stdin, Stdout, Stderr](https://pkg.go.dev/os#pkg-variables)
- [testing/iotest.ErrReader](https://pkg.go.dev/testing/iotest#ErrReader)
- [Go by Example: Line Filters](https://gobyexample.com/line-filters)
- [Go by Example: Exit](https://gobyexample.com/exit)
