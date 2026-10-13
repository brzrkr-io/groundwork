## Task 1: where should the slice end?

`len(lines)-1` is the index of the last line, and the end bound of `s[a:b]` is
*not* included. To keep everything up to the end, leave the end bound out:
`lines[1:]`.

## Task 2: start one later

`hosts[i:]` starts *at* the drained host. The hosts after it start at index
`i+1`.

## Task 3: check before you index

When `hosts` is empty, `len(hosts)-1` is `-1`, which is never a valid index.
Check `len(hosts) == 0` first and return `""` early.

## Almost the answer

```go
return lines[1:] // task 1
```

```go
after = hosts[i+1:] // task 2
```

```go
if len(hosts) == 0 { // task 3
	return ""
}
return hosts[len(hosts)-1]
```
