# Sorting: fix the node report

A teammate wrote a small package that orders Kubernetes nodes for a capacity
report. It compiles and looks right, but each function has one sorting bug.

**New here:** nothing; practice finding sorting bugs.

## Learn

### Recap: the sorting tools

| Tool | What it does |
|---|---|
| `slices.Sort(s)` | sorts `s` *in place*: smallest or alphabetical first; returns nothing |
| `slices.Clone(s)` | returns a new slice with the same values |
| `slices.SortFunc(s, fn)` | sorts `s` in place using your comparison `fn` |
| `cmp.Compare(x, y)` | `-1` if `x < y`, `0` if equal, `+1` if `x > y` |

In a comparison function, negative means `a` goes first, positive means `b`
goes first, zero means "equal".

### Three classic sorting bugs

1. **Sorting someone else's slice.** A slice you were given shares its data
   with the caller. Sorting it in place reorders theirs too.
2. **Backwards order.** `cmp.Compare(x, y)` sorts smallest first. Swapping
   which value goes first and which goes second flips the order.
3. **No tie-break.** When the comparison returns `0`, `SortFunc` may put the
   two elements in any order. Add a second comparison for that case.

A tie-break looks like this, here for words sorted by length, then
alphabetically:

```go
slices.SortFunc(words, func(a, b string) int {
	if d := cmp.Compare(len(a), len(b)); d != 0 {
		return d // the lengths differ, so they decide
	}
	return cmp.Compare(a, b) // same length: alphabetical
})
```

> **Why is it written like this?** `if d := ...; d != 0` is Go's `if` with a
> short statement first; the `;` is required. `d` only exists inside the `if`.
> The name `d` is the author's choice.

## Your task

Each function below has exactly one bug. Read the code, find the bug and fix
it. You don't need to rewrite anything from scratch.

### 1. Stop reordering the caller's slice

**Goal:** make `SortedNodes` return the names in alphabetical order **without** changing the slice it was given.

Right now the returned slice is correct, but the caller's slice gets sorted
too.

```go
nodes := []string{"node-c", "node-a", "node-b"}
SortedNodes(nodes)
// => ["node-a" "node-b" "node-c"]
// nodes is still ["node-c" "node-a" "node-b"]
```

**Think it through:**
- Which slice does `slices.Sort` get right now: the caller's, or a copy?
- Which tool from the recap table makes a copy? Remember to sort and return the copy.

### 2. Put the fullest disks first

**Goal:** make `FullestFirst` return node names with the **highest** disk usage first.

`FullestFirst` gets a map from node name to disk usage in percent. Nodes with
the same usage stay in alphabetical order. Right now the emptiest disks come
first.

```go
FullestFirst(map[string]int{"node-a": 40, "node-b": 91, "node-c": 40})
// => ["node-b" "node-a" "node-c"]
```

**Think it through:**
- Which of the three classic bugs is this?
- Only the usage comparison is wrong. Leave the alphabetical tie-break line alone.

### 3. Make ties predictable

**Goal:** make `LeastBusy` put nodes with the same pod count in alphabetical order.

`LeastBusy(nodes, pods)` gets a slice of node names and a map from node name
to the number of pods on it. It returns a sorted **copy** of the names:
**fewest** pods first, ties alphabetical. The caller's slice is not changed.
Right now tied nodes come out in whatever order they went in.

```go
LeastBusy(
	[]string{"node-c", "node-b", "node-a"},
	map[string]int{"node-a": 12, "node-b": 5, "node-c": 5},
)
// => ["node-b" "node-c" "node-a"]
```

**Think it through:**
- What does the comparison return when two nodes have the same pod count?
- Look at how `FullestFirst` handles ties in the same file.

## Learn more

- [Package slices: SortFunc](https://pkg.go.dev/slices#SortFunc)
- [Package cmp: Compare](https://pkg.go.dev/cmp#Compare)
- [Kubernetes: Nodes](https://kubernetes.io/docs/concepts/architecture/nodes/)
