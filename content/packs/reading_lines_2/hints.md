## Task 1: Fix FindPanics

Look at where `lineNo++` sits: the check runs before the counter goes up, so
the first line is reported as `0`. Also look at what the function returns
after the loop.

### Closer

Move `lineNo++` to the top of the loop. After the loop, check `sc.Err()` and
return `nil, err` when it isn't `nil`, instead of always returning `nil`.

### Answer

```go
for sc.Scan() {
	lineNo++
	if strings.Contains(sc.Text(), "panic:") {
		found = append(found, lineNo)
	}
}
if err := sc.Err(); err != nil {
	return nil, err
}
return found, nil
```

## Task 2: Finish Grep

`Grep` needs its own line counter, just like `FindPanics`. For each line,
check whether it contains `word` and, if so, format it with its number.

### Closer

Declare `lineNo := 0` before the loop and increase it at the top. Use
`strings.Contains(line, word)` and `fmt.Sprintf("%d: %s", lineNo, line)`.
Add `import "fmt"`.

### Answer

```go
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

## Task 3: Handle huge lines

The scanner's default line limit is 64 KiB. One call raises it, and it has to
happen after the scanner is created but before the first `sc.Scan()`.

### Closer

Call `sc.Buffer(...)` right after `bufio.NewScanner(r)`. The first argument is
a starting buffer from `make`; the second is the maximum, `1024*1024`.

### Answer

```go
sc := bufio.NewScanner(r)
sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
```

A 2 MiB line still fails, and the existing `sc.Err()` check returns
`bufio.ErrTooLong`.
