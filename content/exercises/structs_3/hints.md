## Task 1: count the fields first

`strings.Fields(line)` gives you a `[]string`. Check `len(fields) != 3`
before you touch `fields[0]`, and return `Host{}` with an error from
`fmt.Errorf` or `errors.New`. Otherwise build the `Host` from `fields[0]`,
`fields[1]` and `fields[2]`.

## Task 2: the line number is the index plus one

`range` over `strings.Split(text, "\n")` gives you the index `i` of each line,
starting at 0. The line number is `i+1`. Use `continue` to skip blank lines
and comments.

## Task 2: near-solution

```go
var hosts []Host
for i, line := range strings.Split(text, "\n") {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		continue
	}
	h, err := ParseHost(line)
	if err != nil {
		return nil, fmt.Errorf("line %d: %v", i+1, err)
	}
	hosts = append(hosts, h)
}
return hosts, nil
```

## Task 3: make the map first

Create `byRegion := map[string]int{}` before the loop, add one with
`byRegion[h.Region]++` for each host, then return
`Summary{Total: len(hosts), ByRegion: byRegion}`.
