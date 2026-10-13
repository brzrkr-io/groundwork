## Task 1: check Replace first

Inside the `if dep.Path == path` block, return `dep.Replace.Version` when
`dep.Replace != nil`, and `dep.Version` otherwise.

## Task 2: parse, then format

```go
t, err := time.Parse(time.RFC3339, s.Value)
if err != nil {
	return "unknown"
}
return t.UTC().Format("2006-01-02 15:04") + " UTC"
```

Run this when you find the setting with `s.Key == "vcs.time"`. Check
`info == nil` first.

## Task 3: append in a loop

```go
var lines []string
for _, path := range paths {
	lines = append(lines, path+" "+DepVersion(info, path))
}
return lines
```
