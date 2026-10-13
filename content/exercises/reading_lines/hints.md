## Task 1: the scanner loop

Make a scanner with `sc := bufio.NewScanner(r)`, then loop with
`for sc.Scan() { ... }`. Add one to a counter inside the loop. After the loop,
`return count, sc.Err()` covers both the success and the error case.

## Task 2: trim, skip, append

Inside the loop, trim first: `line := strings.TrimSpace(sc.Text())`. Then
`continue` if `line == ""` or `strings.HasPrefix(line, "#")`. Otherwise
`append` it. After the loop, return `nil, err` if `sc.Err()` is not `nil`.

## Task 3: count before you skip

Put `lineNo++` as the very first statement in the loop, before the
`continue`, so skipped lines still count. Then split with
`strings.Cut(line, ":")` and convert the port with `strconv.Atoi`.

## Task 3: almost the answer

```go
host, portText, found := strings.Cut(line, ":")
if !found {
	return nil, fmt.Errorf("line %d: missing port in %q", lineNo, line)
}
port, err := strconv.Atoi(portText)
if err != nil || port < 1 || port > 65535 {
	return nil, fmt.Errorf("line %d: invalid port %q", lineNo, portText)
}
targets = append(targets, Target{Host: host, Port: port})
```
