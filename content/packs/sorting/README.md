# Sorting

Reports are easier to read in a predictable order: hosts alphabetically, the
noisiest services first. Go's `slices` package sorts for you. You only decide
what "comes first" means.

**New here:** `slices.Sort`, `slices.SortFunc`, `cmp.Compare`, `slices.Sorted` with `maps.Keys`.

## Learn

### Sorting a slice: `slices.Sort`

```go
import "slices"

scores := []int{40, 7, 19}
slices.Sort(scores)
// scores is now [7 19 40]
```

`slices.Sort` puts numbers smallest first and strings in alphabetical order.
It sorts *in place*: it returns nothing and rearranges the slice you hand it.

> **Why is it written like this?** Because it returns nothing, Go rejects
> `x := slices.Sort(s)`. Call it on its own line, then use the slice.

### Leave the caller's slice alone

A slice passed to a function still points at the caller's data. If you sort
it, you reorder the caller's slice too, which is a nasty surprise for them.

So sort a *copy*. `slices.Clone(s)` (from *append and copy*) returns a new
slice with the same values. Changes to the clone don't touch `s`:

```go
original := []int{3, 1, 2}
mine := slices.Clone(original)
mine[0] = 99
// original is still [3 1 2]
```

### Your own order: `slices.SortFunc` and `cmp.Compare`

`slices.SortFunc(s, fn)` takes a slice and a *comparison function*. The
function gets two elements, `a` and `b`, and returns an `int`:

| Return value | Meaning |
|---|---|
| negative (like `-1`) | `a` goes before `b` |
| zero | `a` and `b` count as equal |
| positive (like `1`) | `a` goes after `b` |

You rarely work out that number yourself. `cmp.Compare(x, y)` from the `cmp`
package returns `-1` if `x < y`, `0` if equal, `+1` if `x > y`. Here, fruit
sorted by price, cheapest first:

```go
import (
	"cmp"
	"slices"
)

price := map[string]int{"kiwi": 4, "fig": 9, "plum": 4}
fruit := []string{"fig", "kiwi", "plum"}

slices.SortFunc(fruit, func(a, b string) int {
	return cmp.Compare(price[a], price[b])
})
```

To **reverse** the order (dearest first), swap the arguments you give
`cmp.Compare`.

> **Why is it written like this?** Go requires the comparison function to
> take two values of the slice's element type and return `int`; that's the
> shape `SortFunc` expects. The names `a` and `b` are convention. Writing the
> function right inside the call (a *function literal*) is common, not required.

### Breaking ties

`kiwi` and `plum` both cost 4, so the comparison returns `0` for them.
`SortFunc` may then put them in either order, so your output isn't
predictable. Fix it with a *tie-break*: when the first comparison says
"equal", compare something else.

The pattern uses an `if` with a short statement in front:

```go
if d := len(a) - len(b); d != 0 {
	return d // lengths differ: they decide
}
// lengths are equal: fall through to the next comparison
```

`if d := ...; d != 0` stores a value in `d` and then checks it. `d` only
exists inside that `if`.

### Sorted map keys

As you saw in *Maps*, ranging over a map visits keys in random order.
Two functions together give you sorted keys:

- `maps.Keys(m)` hands out the keys of `m` one at a time (it is not a slice).
- `slices.Sorted(x)` collects those keys into a new slice, sorts it and
  returns it.

Pass the result of one straight into the other; the inner call runs first.
The map isn't changed. An empty map gives an empty slice.

## Your task

Your monitoring tool collects host names and error counts per service, and the
on-call engineer wants them in a sensible order.

### 1. Sort hosts without touching the input

**Goal:** make `SortedHosts` return a **new** slice with the hosts in alphabetical order.

The `hosts` slice the caller passed in must keep its original order. An empty
slice gives an empty slice.

```go
hosts := []string{"web-2", "db-1", "web-1"}
SortedHosts(hosts)
// => ["db-1" "web-1" "web-2"]
// hosts is still ["web-2" "db-1" "web-1"]
```

**Think it through:**
- Right now it returns `hosts` itself. What do you need first so you don't sort the caller's slice?
- `slices.Sort` returns nothing. Which name do you `return` after sorting?

### 2. List service names in order

**Goal:** make `ServiceNames(errs)` return the map's keys (the service names) in alphabetical order.

Add `"maps"` to the imports at the top of the file. An empty map gives an
empty slice.

```go
ServiceNames(map[string]int{"web": 3, "api": 9, "db": 0})
// => ["api" "db" "web"]
```

**Think it through:**
- The values (error counts) don't matter here, only the keys.
- Which two functions from *Sorted map keys* do you need, and which goes inside which?

### 3. Rank services by errors

**Goal:** finish the comparison function in `RankByErrors` so **more** errors come first and ties are alphabetical.

`RankByErrors` already gets the names in alphabetical order from
`ServiceNames` and calls `slices.SortFunc`. You only fill in the body of the
comparison function. Add `"cmp"` to the imports.

```go
RankByErrors(map[string]int{"web": 3, "api": 9, "db": 3, "cache": 0})
// => ["api" "db" "web" "cache"]
```

**Think it through:**
- `a` and `b` are service *names*. Where do you look up their error counts?
- "More first" is the reverse of what `cmp.Compare` does by default. What do you swap?
- When the counts are equal, compare the names themselves.

## Learn more

- [Package slices: Sort, SortFunc, Sorted](https://pkg.go.dev/slices)
- [Package cmp: Compare](https://pkg.go.dev/cmp#Compare)
- [Package maps: Keys](https://pkg.go.dev/maps#Keys)
- [Go by Example: Sorting by Functions](https://gobyexample.com/sorting-by-functions)
