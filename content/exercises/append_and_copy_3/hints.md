## Task 1: Build the deploy plan

Start with an empty result and two loops: first over `canaries`, then over
`hosts`. Only keep names that aren't blank.

### Closer

Start with `var plan []string`. Inside each loop, check `canaries[i] != ""`
before you `append` it. A `nil` slice is fine when nothing is left.

```go
for i := range len(canaries) {
	if canaries[i] != "" {
		plan = append(plan, ...)
	}
}
```

### Answer

```go
var plan []string
for i := range len(canaries) {
	if canaries[i] != "" {
		plan = append(plan, canaries[i])
	}
}
for i := range len(hosts) {
	if hosts[i] != "" {
		plan = append(plan, hosts[i])
	}
}
return plan
```

Canaries are appended first, so they come first in the plan.

## Task 2: Split it into batches

Walk `hosts` in steps of `size`, and take one slice of hosts per step. Check
`size` first: if it's less than 1, the loop would never move forward.

### Closer

Return `nil` when `size < 1`. Loop with
`for start := 0; start < len(hosts); start += size`, and cut each batch with
`hosts[start:end]`, where `end := min(start+size, len(hosts))`. That slice
shares elements with `hosts`, so copy it with `slices.Clone(...)` (add
`import "slices"`) before you append it.

### Answer

```go
if size < 1 {
	return nil
}
var batches [][]string
for start := 0; start < len(hosts); start += size {
	end := min(start+size, len(hosts))
	batches = append(batches, slices.Clone(hosts[start:end]))
}
return batches
```

`min` keeps the last batch from running past the end, and `slices.Clone`
makes each batch its own copy.
