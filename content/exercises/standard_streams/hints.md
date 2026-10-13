## Task 1: three small changes

Change the condition to `len(args) != 1`, write to `stderr` instead of
`stdout`, and return `2` instead of `0`.

## Task 2: the scanner loop

```go
sc := bufio.NewScanner(stdin)
for sc.Scan() {
	host := strings.TrimSpace(sc.Text())
	// skip blank and # lines with continue
	// print with fmt.Fprintln(stdout, host) if strings.HasPrefix(host, prefix)
}
return 0
```

## Task 3: remember, don't return

Keep a `lineNo` counter (add one at the top of the loop) and a `code := 0`
variable. For an invalid host, print the message, set `code = 1` and
`continue`. After the loop:

```go
if err := sc.Err(); err != nil {
	fmt.Fprintf(stderr, "hostfilter: reading stdin: %v\n", err)
	return 1
}
return code
```
