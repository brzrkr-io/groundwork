## Task 1: build the pattern piece by piece

Write down what a tag is made of: start, `v`, number, dot, number, dot,
number, end. Each piece has an entry in the README's table: `^`, `v`, `\d+`,
`\.`, `$`.

## Task 1: the pattern

```go
var semverTag = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
```

`IsSemverTag` already calls `semverTag.MatchString(tag)`, so the pattern is
the only change.

## Task 2: groups and Atoi

Put parentheses around each `\d+`. After `m := semverTag.FindStringSubmatch(tag)`
and a `nil` check, the numbers are in `m[1]`, `m[2]` and `m[3]`. Pass each one
to `strconv.Atoi` and check its error.

## Task 3: almost the answer

```go
m := logLine.FindStringSubmatch(line)
if m == nil {
	return LogLine{}, fmt.Errorf("not a log line: %q", line)
}
return LogLine{
	Time:  m[logLine.SubexpIndex("time")],
	Level: m[logLine.SubexpIndex("level")],
	// Service and Message work the same way.
}, nil
```
