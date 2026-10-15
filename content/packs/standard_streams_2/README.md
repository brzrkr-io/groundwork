# Standard streams: grep-style exit codes

Scripts often ask a yes/no question by running a search tool and checking its
exit code. Here you fix a log scanner so scripts can trust both its output and
its answer.

**New here:** grep-style exit codes (`0` / `1` / `2`). The rest is practice.

## Learn

### Recap

- Results go to *stdout*. Messages for humans go to *stderr*.
- Only stdout flows into a pipe or a `>` file.
- `Run` gets the streams as parameters (`stdin`, `stdout`, `stderr`), so the
  tests can pass fake ones. Its return value is the exit code.

### grep's three exit codes

`grep` uses its exit code to answer a question, not only to say "it worked":

| Code | Meaning |
|---|---|
| `0` | at least one line matched |
| `1` | nothing matched (not an error, just "no") |
| `2` | something went wrong: bad arguments, unreadable input |

A script can branch on the answer directly:

```bash
if logscan ERROR < app.log > errors.txt; then
	echo "errors found, paging on-call"
fi
```

`2` is kept apart from `1` on purpose. "The log has no errors" and "I could
not read the log" must never look the same, or a broken disk would look like
a healthy service.

Gotcha: under `set -e`, a script stops when a tool returns `1`. If "no match"
is fine, write `logscan ERROR < app.log || true`.

### Choosing the stream with Fprintf

`fmt.Fprintf` and `fmt.Fprintln` take the writer to print to as their
**first** argument:

```go
fmt.Fprintln(stdout, "web01")             // a result
fmt.Fprintf(stderr, "checked %d hosts\n", n) // a note for the human
```

> **Why is it written like this?** Go requires the writer to come first; the
> `F` in `Fprintf` means "to this writer". Which stream you pick is your
> choice, and the convention is: data on stdout, everything else on stderr.

A summary like `logscan: 2 matching line(s)` is a note, not a matched line. On
stdout it would land in `errors.txt` or in the next command of a pipe.

### Checking a Scanner for errors

`sc.Scan()` returns `false` both at the end of input **and** when reading
fails. After the loop, `sc.Err()` tells you which: `nil` means "just the end".
Check it with an `if` and a short statement, shown here with a different
call:

```go
if err := cfg.Validate(); err != nil {
	fmt.Fprintf(stderr, "config: %v\n", err)
	return 2
}
```

> **Why is it written like this?** The `if x := ...; cond` form is Go syntax:
> it declares `err` only inside the `if`. Using it is convention. `%v` prints
> the error's message.

## Your task

`logscan PATTERN` copies every line of stdin that contains `PATTERN` to
stdout. Matching uses `strings.Contains`, so it is case-sensitive: `error`
does not match `ERROR`. The usage check already works (`usage: logscan
PATTERN` on stderr, exit `2`). The rest has bugs.

### 1. Move the summary to stderr

**Goal:** print the summary line to `stderr` instead of `stdout`.

Only matching lines stay on stdout. The summary text itself doesn't change:

```text
$ logscan ERROR < app.log 2>/dev/null
2026-10-11T09:00:07Z ERROR db connection refused
2026-10-11T09:00:12Z ERROR db connection refused
$ logscan ERROR < app.log >/dev/null
logscan: 2 matching line(s)
```

With no matches the summary is `logscan: 0 matching line(s)`.

**Think it through:**
- Find the line that prints the summary. Which writer does it get first?
- `Run` was given two writers. Which one is for notes?

### 2. Return grep-style exit codes

**Goal:** `Run` returns `0` if at least one line matched, `1` if none did.

Empty input counts as "none matched", so it returns `1`.

```text
$ logscan PANIC < app.log; echo $?
logscan: 0 matching line(s)
1
```

**Think it through:**
- Which variable already counts the matched lines?
- Replace the fixed `return 0` with a decision based on that count.

### 3. Report read errors

**Goal:** if reading stdin fails, print an error to `stderr` and return `2`.

The message is `logscan: reading stdin: ` followed by the error (`%v`) and a
newline:

```text
logscan: reading stdin: disk read failed
```

- Lines matched before the failure stay on stdout.
- Do **not** print the summary in this case; the count would be wrong.
- The tests make stdin fail with `iotest.ErrReader`, as in *Standard streams:
  stdin, stdout, stderr and exit codes*.

**Think it through:**
- Where does the `TODO(3)` sit: before or after the summary? Why does that
  matter?
- Which method of `sc` reports a read error?

## Learn more

- [GNU grep: Exit status](https://www.gnu.org/software/grep/manual/grep.html#Exit-Status)
- [bufio.Scanner](https://pkg.go.dev/bufio#Scanner)
- [os.Exit](https://pkg.go.dev/os#Exit)
- [testing/iotest.ErrReader](https://pkg.go.dev/testing/iotest#ErrReader)
