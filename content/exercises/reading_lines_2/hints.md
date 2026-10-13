## Task 1: where does the counter move?

The counter goes up *after* the check, so the first line is reported while
`lineNo` is still `0`. Move `lineNo++` to the top of the loop. Then check
`sc.Err()` after the loop instead of always returning `nil`.

## Task 2: count, check, format

Grep needs its own `lineNo` counter, declared before the loop and increased
at the top of it. Use `strings.Contains(line, word)` and
`fmt.Sprintf("%d: %s", lineNo, line)`.

## Task 3: one line before the loop

`sc.Buffer` must run after `bufio.NewScanner` and before the first
`sc.Scan()`. Copy the call from the Learn section.

## Almost the answer

```go
sc := bufio.NewScanner(r)
sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
var matches []string
lineNo := 0
for sc.Scan() {
	lineNo++
	line := sc.Text()
	if strings.Contains(line, word) {
		matches = append(matches, fmt.Sprintf("%d: %s", lineNo, line))
	}
}
```
