## Task 1: Move the summary to stderr

Find the line that prints the summary. Look at which writer it sends its text to: the matching lines and the summary shouldn't share one.

### Closer

`fmt.Fprintf` takes the writer as its first argument. Swap `stdout` for the other writer `Run` was given.

### Answer

```go
fmt.Fprintf(stderr, "logscan: %d matching line(s)\n", matches)
```

## Task 2: Return grep-style exit codes

`matches` already counts the matched lines. Decide the return value from that count instead of always returning `0`.

### Closer

Replace `return 0` with an `if` on `matches == 0`: one branch returns `1`, the other `0`.

### Answer

```go
if matches == 0 {
	return 1
}
return 0
```

## Task 3: Report read errors

After the scan loop, `sc.Err()` tells you whether reading stopped because of an error. Check it before printing the summary, so a failed read never reaches the summary.

### Closer

Put an `if err := sc.Err(); err != nil { ... }` right after the loop. Inside, print with `fmt.Fprintf(stderr, ...)` using `%v` for the error, then return early.

### Answer

```go
if err := sc.Err(); err != nil {
	fmt.Fprintf(stderr, "logscan: reading stdin: %v\n", err)
	return 2
}
```

Returning here skips the summary, and `2` keeps "could not read" apart from "no match".
