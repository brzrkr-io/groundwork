# Standard streams: grep-style exit codes

Scripts often ask a yes/no question by running a search tool and checking its
exit code. Here you fix a log scanner so scripts can trust both its output and
its answer.

## Learn

**Recap.** Results go to stdout, diagnostics go to stderr, and only stdout
flows through a pipe. `Run` takes the streams as arguments, so `main` passes
`os.Stdin`, `os.Stdout` and `os.Stderr`, and the tests pass a
`strings.NewReader` and `strings.Builder`s. The exit code tells the shell
whether the tool worked.

### grep's three exit codes

`grep` uses its exit code to answer a question, not only to report success:

| Code | grep's meaning |
|---|---|
| `0` | at least one line matched |
| `1` | nothing matched (not an error, just "no") |
| `2` | something went wrong: bad arguments, unreadable input |

That lets a script branch on the answer directly:

```bash
if logscan ERROR < app.log > errors.txt; then
	echo "errors found, paging on-call"
fi
```

`2` is kept apart from `1` on purpose. "The log has no errors" and "I could
not read the log" must never look the same. Otherwise a broken disk would
report a healthy service.

One gotcha: under `set -e`, a script stops when such a tool returns `1`. When
"no match" is fine, write `logscan ERROR < app.log || true`.

### Summaries belong on stderr

A count like `logscan: 2 matching line(s)` is useful for a human, but it is
not one of the matched lines. On stdout it would end up in `errors.txt` above,
or in the next command of a pipe. On stderr, it shows in the terminal and
stays out of the data.

## Your task

`logscan PATTERN` copies every line of stdin that contains `PATTERN` to
stdout. Matching uses `strings.Contains`, so it is case-sensitive: `error`
does not match `ERROR`. The usage check already works (`usage: logscan
PATTERN` on stderr, exit `2`). The rest has bugs.

### 1. Move the summary to stderr

After reading all of stdin, `Run` prints a summary, but to stdout. Print it to
stderr instead, so stdout carries only matching lines:

```text
$ logscan ERROR < app.log 2>/dev/null
2026-10-11T09:00:07Z ERROR db connection refused
2026-10-11T09:00:12Z ERROR db connection refused
$ logscan ERROR < app.log >/dev/null
logscan: 2 matching line(s)
```

With no matches the summary is `logscan: 0 matching line(s)`.

### 2. Return grep-style exit codes

`Run` always returns `0`. Return `0` if at least one line matched and `1` if
none did (empty input included).

```text
$ logscan PANIC < app.log; echo $?
logscan: 0 matching line(s)
1
```

### 3. Report read errors

If `sc.Err()` returns an error, print the following to stderr and return `2`:

```text
logscan: reading stdin: disk read failed
```

The part after `reading stdin: ` is the error (`%v`). Lines matched before the
failure stay on stdout. Do **not** print the summary in this case: the count
would be wrong. The tests make stdin fail with `iotest.ErrReader`, as in
*Standard streams: stdin, stdout, stderr and exit codes*.

## Learn more

- [GNU grep: Exit status](https://www.gnu.org/software/grep/manual/grep.html#Exit-Status)
- [os.Exit](https://pkg.go.dev/os#Exit)
- [testing/iotest.ErrReader](https://pkg.go.dev/testing/iotest#ErrReader)
