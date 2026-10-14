## Task 1: Parse one line

Split the line into its words, check that there are exactly three, then build a `Host` from them. You'll need to import `strings` and `fmt`.

### Closer

`fields := strings.Fields(line)` gives you a `[]string`. Check `len(fields) != 3` before you touch `fields[0]`, and in that case return `Host{}` with an error from `fmt.Errorf(...)`.

### Answer

```go
fields := strings.Fields(line)
if len(fields) != 3 {
	return Host{}, fmt.Errorf("want 3 fields (name ip region), got %d in %q", len(fields), line)
}
return Host{Name: fields[0], IP: fields[1], Region: fields[2]}, nil
```

## Task 2: Parse the whole inventory

Go through the text line by line, skip blank lines and comments, and hand each other line to `ParseHost`. The line number in the error is the loop index plus one.

### Closer

`for i, line := range strings.Split(text, "\n")` gives you each line and its index. Use `strings.TrimSpace`, then `continue` when the line is `""` or `strings.HasPrefix(line, "#")`. Wrap errors with `fmt.Errorf("line %d: %v", i+1, err)`.

### Answer

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

## Task 3: Summarise the hosts

`Total` is just how many hosts there are. For `ByRegion`, make a map before the loop and count each host's region.

### Closer

Create `byRegion := map[string]int{}`, then add one per host with `byRegion[h.Region]++`. Finish with a `Summary{...}` literal using `len(hosts)`.

### Answer

```go
byRegion := map[string]int{}
for _, h := range hosts {
	byRegion[h.Region]++
}
return Summary{Total: len(hosts), ByRegion: byRegion}
```
