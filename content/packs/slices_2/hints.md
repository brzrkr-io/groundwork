## Task 1: Skip the header, keep every host

The header is at index `0`, so the hosts start at index `1`. Look at the end
bound: `len(lines)-1` is the index of the last host, and the end bound of
`s[a:b]` is *not* included.

### Closer

Leaving out the end bound runs the slice all the way to the end: `lines[1:]`
keeps index `1` and everything after it.

### Answer

```go
return lines[1:]
```

With no end bound, the slice keeps every line after the header, including the
last one.

## Task 2: Drain one host

`hosts[i:]` starts *at* the drained host, so it is still in `after`. Where do
the hosts after index `i` start?

### Closer

The next host is at index `i+1`. Change only the start bound of the `after`
line: `hosts[... :]`.

### Answer

```go
before = hosts[:i]
after = hosts[i+1:]
```

`hosts[:i]` stops just before `i` and `hosts[i+1:]` starts just after it, so
`hosts[i]` is in neither.

## Task 3: Don't crash on an empty inventory

When `hosts` is empty, `len(hosts)-1` is `-1`, which is never a valid index.
Check for the empty case before you index.

### Closer

Test the length first and return early:

```go
if len(hosts) == 0 {
	// return the empty string here
}
```

### Answer

```go
if len(hosts) == 0 {
	return ""
}
return hosts[len(hosts)-1]
```

A `nil` slice has length `0` too, so this one check covers both the empty and
the `nil` inventory.
