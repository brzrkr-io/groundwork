## Task 1: Recognise release tags

Write down what a tag is made of, in order: start, `v`, number, dot, number,
dot, number, end. Each piece has an entry in the README's table.

### Closer

The pieces are `^`, `v`, `\d+`, `\.` and `$`. Write the pattern in backticks:
`` `^v\d+\.` `` covers the first number and dot; repeat for the rest.

### Answer

```go
var semverTag = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
```

`^` and `$` stop anything sneaking in before or after, and `\.` is a real dot.

## Task 2: Parse release tags

First wrap each `\d+` in parentheses so the numbers are captured. Then
`FindStringSubmatch` hands you the whole match plus one string per group.

### Closer

After `m := semverTag.FindStringSubmatch(tag)` and a `nil` check, the numbers
are in `m[1]`, `m[2]` and `m[3]`. Turn each into an `int` with
`n, err := strconv.Atoi(m[1])`, and add `"strconv"` to the imports.

### Answer

```go
var semverTag = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

func ParseSemver(tag string) (Version, bool) {
	m := semverTag.FindStringSubmatch(tag)
	if m == nil {
		return Version{}, false
	}
	var nums [3]int
	for i, s := range m[1:] {
		n, err := strconv.Atoi(s)
		if err != nil {
			return Version{}, false
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2]}, true
}
```

`m[0]` is the whole tag, so `m[1:]` is exactly the three captured numbers.

## Task 3: Parse log lines

`FindStringSubmatch` returns `nil` when the line doesn't match; handle that
first. Otherwise each named group sits at the index `SubexpIndex` gives you.

### Closer

The error is `fmt.Errorf("not a log line: %q", line)`. A field looks like
`Time: m[logLine.SubexpIndex("time")]`.

### Answer

```go
m := logLine.FindStringSubmatch(line)
if m == nil {
	return LogLine{}, fmt.Errorf("not a log line: %q", line)
}
return LogLine{
	Time:    m[logLine.SubexpIndex("time")],
	Level:   m[logLine.SubexpIndex("level")],
	Service: m[logLine.SubexpIndex("service")],
	Message: m[logLine.SubexpIndex("message")],
}, nil
```

Looking groups up by name means you never have to count parentheses.
