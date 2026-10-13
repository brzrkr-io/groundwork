## Task 1: two loops, one result

Start with `var plan []string`. Loop over `canaries` by index and append each
name that isn't `""`. Then do the same for `hosts`.

## Task 2: step by size

Guard against `size < 1` first, or the loop never ends. Then loop with
`for start := 0; start < len(hosts); start += size`. Each batch is
`hosts[start:end]` where `end` is `start + size`, but no more than `len(hosts)`:
`end := min(start+size, len(hosts))`.

## Task 2: copy each batch

`hosts[start:end]` shares its elements with `hosts`. Wrap it in
`slices.Clone(...)` (and add `import "slices"`) before appending it to
`batches`.

## Almost the answer

```go
var batches [][]string
for start := 0; start < len(hosts); start += size {
	end := min(start+size, len(hosts))
	batches = append(batches, slices.Clone(hosts[start:end]))
}
return batches
```
