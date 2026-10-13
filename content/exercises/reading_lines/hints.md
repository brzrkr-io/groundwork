## Task 1: Count the lines

Wrap the reader in a `bufio.Scanner` and loop once per line, adding one to a
counter each time. After the loop, the scanner can tell you whether reading
failed.

### Closer

`sc := bufio.NewScanner(r)`, then `for sc.Scan() { count++ }`. After the loop,
`sc.Err()` is `nil` on success or the read error.

### Answer

```go
sc := bufio.NewScanner(r)
count := 0
for sc.Scan() {
	count++
}
return count, sc.Err()
```

## Task 2: Read the hostnames

Use the same scanner loop, but clean each line before you look at it. Skip
blank lines and comments; keep everything else.

### Closer

Inside the loop: `line := strings.TrimSpace(sc.Text())`, then `continue` if
`line == ""` or `strings.HasPrefix(line, "#")`. After the loop, return
`nil, err` when `sc.Err()` is not `nil`.

### Answer

```go
sc := bufio.NewScanner(r)
var hosts []string
for sc.Scan() {
	line := strings.TrimSpace(sc.Text())
	if line == "" || strings.HasPrefix(line, "#") {
		continue
	}
	hosts = append(hosts, line)
}
if err := sc.Err(); err != nil {
	return nil, err
}
return hosts, nil
```

## Task 3: Parse host:port targets

Start from your task 2 loop and add a line counter. Count before you skip, so
blank and comment lines still count. Then split each kept line and convert the
port.

### Closer

Put `lineNo++` as the first statement in the loop. Split with
`host, portText, found := strings.Cut(line, ":")` and convert with
`strconv.Atoi(portText)`. Build errors like
`fmt.Errorf("line %d: invalid port %q", lineNo, portText)`.

### Answer

```go
for sc.Scan() {
	lineNo++ // count every line, even the ones we skip
	line := strings.TrimSpace(sc.Text())
	if line == "" || strings.HasPrefix(line, "#") {
		continue
	}
	host, portText, found := strings.Cut(line, ":")
	if !found {
		return nil, fmt.Errorf("line %d: missing port in %q", lineNo, line)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("line %d: invalid port %q", lineNo, portText)
	}
	targets = append(targets, Target{Host: host, Port: port})
}
```

Declare `sc`, `var targets []Target` and `lineNo := 0` before the loop, and
check `sc.Err()` after it as in task 2.
