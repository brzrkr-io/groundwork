## Task 1: Fix the usage check

Three things are wrong in the usage branch: the condition lets extra arguments
through, the line goes to the wrong stream, and the exit code says success.

### Closer

Change the condition to `len(args) != 1`, print to `stderr` instead of
`stdout`, and return `2` instead of `0`.

### Answer

```go
if len(args) != 1 {
	fmt.Fprintln(stderr, "usage: hostfilter PREFIX")
	return 2
}
```

## Task 2: Filter the hosts

Read stdin one line at a time with a `bufio.Scanner`. Trim each line, skip the
blank and comment ones, and print the ones that match the prefix.

### Closer

`sc := bufio.NewScanner(stdin)` and `for sc.Scan()`, with
`strings.TrimSpace(sc.Text())`. Skip with `continue` when the host is `""` or
`strings.HasPrefix(host, "#")`. Add `"bufio"` and `"strings"` to the imports.

### Answer

```go
sc := bufio.NewScanner(stdin)
for sc.Scan() {
	host := strings.TrimSpace(sc.Text())
	if host == "" || strings.HasPrefix(host, "#") {
		continue
	}
	if strings.HasPrefix(host, prefix) {
		fmt.Fprintln(stdout, host)
	}
}
return 0
```

## Task 3: Report problems on stderr

An invalid host shouldn't stop the loop: remember that something went wrong
and keep going. You also need a line counter that counts every line, and a
`sc.Err()` check after the loop.

### Closer

Add `code := 0` and `lineNo := 0` before the loop, and `lineNo++` as the first
line inside it. For an invalid host, print with `%d` and `%q`, set `code = 1`
and `continue`. Return `code` at the end.

### Answer

```go
if strings.ContainsAny(host, " \t") {
	fmt.Fprintf(stderr, "hostfilter: line %d: invalid host %q\n", lineNo, host)
	code = 1
	continue
}

// after the loop:
if err := sc.Err(); err != nil {
	fmt.Fprintf(stderr, "hostfilter: reading stdin: %v\n", err)
	return 1
}
return code
```
