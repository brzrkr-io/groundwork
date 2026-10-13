# Sorting

Reports are easier to read in a predictable order: hosts alphabetically, the
noisiest services first. Go's `slices` package sorts for you. You only decide
what "comes first" means.

## Learn

### Sorting a slice: `slices.Sort`

```go
import "slices"

hosts := []string{"web-2", "db-1", "web-1"}
slices.Sort(hosts)
// hosts is now ["db-1" "web-1" "web-2"]
```

`slices.Sort` puts numbers smallest first and strings in alphabetical order.
It sorts *in place*: it returns nothing and rearranges the slice you hand it.

### Leave the caller's slice alone: `slices.Clone` first

A slice passed to a function still points at the caller's data. If you sort
it, you reorder the caller's slice too, which is a nasty surprise for them.
Sort a copy instead, using `slices.Clone` from *append and copy*:

```go
func Sorted(names []string) []string {
	out := slices.Clone(names) // a new slice with the same values
	slices.Sort(out)           // only out is reordered
	return out
}
```

### Your own order: `slices.SortFunc` and `cmp.Compare`

`slices.SortFunc` takes a slice and a *comparison function*. The function gets
two elements, `a` and `b`, and returns an `int`:

| Return value | Meaning |
|---|---|
| negative (like `-1`) | `a` goes before `b` |
| zero | `a` and `b` count as equal |
| positive (like `1`) | `a` goes after `b` |

You rarely work out that number yourself. `cmp.Compare(x, y)` from the `cmp`
package returns `-1` if `x < y`, `0` if they are equal and `+1` if `x > y`.
Compare the values you care about, in this case the error count of each
service:

```go
import (
	"cmp"
	"slices"
)

errs := map[string]int{"api": 7, "db": 2, "web": 7}
names := []string{"api", "db", "web"}

// Fewest errors first.
slices.SortFunc(names, func(a, b string) int {
	return cmp.Compare(errs[a], errs[b])
})
```

To **reverse** the order, swap the arguments: `cmp.Compare(errs[b], errs[a])`
puts the most errors first.

### Breaking ties

`api` and `web` both have 7 errors, so the comparison above returns `0` for
them. `slices.SortFunc` may then put them in either order, and your report
changes from run to run. Fix that with a *tie-break*: when the first
comparison says "equal", compare something else.

```go
slices.SortFunc(names, func(a, b string) int {
	if c := cmp.Compare(errs[b], errs[a]); c != 0 {
		return c // the counts differ, so they decide
	}
	return cmp.Compare(a, b) // same count: alphabetical
})
// names is now ["api" "web" "db"]
```

`if c := ...; c != 0` stores the comparison in `c` and then checks it, all in
one line. `c` only exists inside that `if`.

### Sorted map keys

As you saw in *Maps*, ranging over a map visits the keys in a random order.
To get the keys in order, write:

```go
import (
	"maps"
	"slices"
)

names := slices.Sorted(maps.Keys(errs))
// ["api" "db" "web"]
```

Read it from the inside out. `maps.Keys(errs)` hands out the keys of the map
one at a time. `slices.Sorted` collects them into a brand new slice, sorts it
and returns it. The map itself is not changed. For an empty map you get an
empty slice.

## Your task

Your monitoring tool collects host names and error counts per service, and the
on-call engineer wants them in a sensible order.

### 1. Sort hosts without touching the input

Change `SortedHosts` to return a **new** slice holding the hosts in
alphabetical order. The `hosts` slice the caller passed in must keep its
original order.

```go
hosts := []string{"web-2", "db-1", "web-1"}
SortedHosts(hosts)
// => ["db-1" "web-1" "web-2"]
// hosts is still ["web-2" "db-1" "web-1"]
```

An empty slice gives an empty slice.

### 2. List service names in order

`ServiceNames` gets a map from service name to error count. Change it to return
the service names (the map keys) in alphabetical order. Add `"maps"` to the
imports at the top of the file.

```go
ServiceNames(map[string]int{"web": 3, "api": 9, "db": 0})
// => ["api" "db" "web"]
```

An empty map gives an empty slice.

### 3. Rank services by errors

`RankByErrors` already gets the names in alphabetical order from
`ServiceNames` and calls `slices.SortFunc`. Finish the comparison function so
that services with **more** errors come first, and services with the **same**
number of errors stay in alphabetical order. Add `"cmp"` to the imports.

```go
RankByErrors(map[string]int{"web": 3, "api": 9, "db": 3, "cache": 0})
// => ["api" "db" "web" "cache"]
```

## Learn more

- [Package slices: Sort, SortFunc, Sorted](https://pkg.go.dev/slices)
- [Package cmp: Compare](https://pkg.go.dev/cmp#Compare)
- [Package maps: Keys](https://pkg.go.dev/maps#Keys)
- [Go by Example: Sorting by Functions](https://gobyexample.com/sorting-by-functions)
