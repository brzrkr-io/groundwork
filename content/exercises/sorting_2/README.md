# Sorting: fix the node report

A teammate wrote a small package that orders Kubernetes nodes for a capacity
report. It compiles and looks right, but each function has one sorting bug.

## Learn

### Recap

`slices.Sort(s)` sorts `s` in place, so clone first with `slices.Clone` if the
slice belongs to someone else. `slices.SortFunc(s, func(a, b T) int {...})`
uses your comparison: negative means `a` first, positive means `b` first, zero
means "equal". `cmp.Compare(x, y)` gives you that number, and swapping its
arguments reverses the order. When two elements compare equal, `SortFunc` may
put them in any order, so add a tie-break.

## Your task

Each function below has exactly one bug. Read the code, find the bug and fix
it. You don't need to rewrite anything from scratch.

### 1. Stop reordering the caller's slice

`SortedNodes` should return the node names in alphabetical order **without**
changing the slice it was given. Right now the returned slice is correct, but
the caller's slice gets sorted too.

```go
nodes := []string{"node-c", "node-a", "node-b"}
SortedNodes(nodes)
// => ["node-a" "node-b" "node-c"]
// nodes is still ["node-c" "node-a" "node-b"]
```

### 2. Put the fullest disks first

`FullestFirst` gets a map from node name to disk usage in percent. It should
return the node names with the **highest** usage first. Nodes with the same
usage are in alphabetical order. Right now the emptiest disks come first.

```go
FullestFirst(map[string]int{"node-a": 40, "node-b": 91, "node-c": 40})
// => ["node-b" "node-a" "node-c"]
```

### 3. Make ties predictable

`LeastBusy` gets a slice of node names and a map from node name to the number
of pods running on it. It returns a sorted copy of the names: **fewest** pods
first, and nodes with the same pod count in alphabetical order. The caller's
slice is not changed. Right now nodes with the same pod count come out in
whatever order they went in.

```go
LeastBusy(
	[]string{"node-c", "node-b", "node-a"},
	map[string]int{"node-a": 12, "node-b": 5, "node-c": 5},
)
// => ["node-b" "node-c" "node-a"]
```

## Learn more

- [Package slices: SortFunc](https://pkg.go.dev/slices#SortFunc)
- [Package cmp: Compare](https://pkg.go.dev/cmp#Compare)
- [Kubernetes: Nodes](https://kubernetes.io/docs/concepts/architecture/nodes/)
