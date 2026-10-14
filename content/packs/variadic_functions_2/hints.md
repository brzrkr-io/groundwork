## Task 1: Fix the alert format

The trailing `", "` comes from writing a separator after every host. Treat the first host differently from the rest, and handle "no hosts" before anything else.

### Closer

Return `severity + ": no hosts"` when `len(hosts) == 0`. Otherwise write `hosts[0]`, then loop over `hosts[1:]` and write `", "` *before* each of those hosts.

### Answer

```go
if len(hosts) == 0 {
	return severity + ": no hosts"
}
var b strings.Builder
b.WriteString(severity)
b.WriteString(": ")
b.WriteString(hosts[0])
for _, h := range hosts[1:] {
	b.WriteString(", ")
	b.WriteString(h)
}
return b.String()
```

Putting the separator before every host except the first means none is left over at the end.

## Task 2: Page about every host

`Alert` accepts any number of hosts, and `down` is already a slice of hosts. You only need one call that passes the whole slice.

### Closer

Put `...` after a slice to spread it into a variadic parameter: `Alert("critical", down...)`. An empty or `nil` slice spreads into zero hosts, so the `if` isn't needed.

### Answer

```go
return Alert("critical", down...)
```

With `nil`, `Alert` gets no hosts and returns `"critical: no hosts"` on its own.

## Task 3: Put the canary first

You can't mix one single value and a spread slice in the same call, so `Alert("warning", canary, others...)` won't compile. Build one slice first.

### Closer

`append([]string{canary}, others...)` makes a slice with `canary` first and then every host in `others`. Then spread that slice into `Alert`.

### Answer

```go
hosts := append([]string{canary}, others...)
return Alert("warning", hosts...)
```

The new slice holds all hosts in the right order, and `hosts...` passes them to `Alert`.
