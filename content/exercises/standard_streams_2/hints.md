## Task 1: one word

The summary line uses `fmt.Fprintf(stdout, ...)`. Which writer should it use?

## Task 2: decide from the count

`matches` already counts the matched lines. Replace `return 0` with an `if`
on `matches == 0`.

## Task 3: check before the summary

Put the error check right after the loop, before the summary, so a read error
returns early:

```go
if err := sc.Err(); err != nil {
	fmt.Fprintf(stderr, "logscan: reading stdin: %v\n", err)
	return 2
}
```
