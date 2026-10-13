## Tasks 1 and 2: slicing does the work

`hosts[:n]` is the first `n` hosts and `hosts[len(hosts)-n:]` is the last
`n`. Both panic when `n` is bigger than `len(hosts)`, so shrink `n` first.

## Shrink n with an if

```go
if n > len(hosts) {
	n = len(hosts)
}
```

After this, `n` is never too big. A `nil` slice has length `0`, so `n` becomes
`0` and `hosts[:0]` is an empty result.

## Task 3: make, then fill

`make([]int, count)` gives you `count` zeros. Loop with `for i := range count`
and set `ports[i]` to `start + i`.

## Almost the answer

```go
func FirstN(hosts []string, n int) []string {
	if n > len(hosts) {
		n = len(hosts)
	}
	return hosts[:n]
}
```

`LastN` is the same, except it returns `hosts[len(hosts)-n:]`.
