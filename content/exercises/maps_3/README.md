# Maps: diff a service catalog

Every platform team ends up with a *service catalog*: which team owns which
service. In this exercise you answer the questions people ask of it during a
reorg, using maps from end to end.

**New here:** nothing; practice with maps, loops and multiple return values.

## Learn

### Recap: the map moves you need

| You want to… | Write | Note |
|---|---|---|
| read a value | `m[k]` | missing key gives the zero value (`""`, `0`) |
| know if a key is there | `v, ok := m[k]` | `ok` is `true` or `false` |
| create a map | `make(map[string]int)` | a `var` map is `nil`; writing to it panics |
| add one to a count | `m[k]++` | a missing key starts at `0` |
| visit every entry | `for k, v := range m` | order is random every run |

Reading a `nil` map is fine: it behaves like an empty one, and a `range` over
it runs zero times.

### Counting with a map

`m[k]++` reads `m[k]` (0 if missing), adds one, and stores it back. So you
never need to "set up" a key before counting it:

```go
votes := make(map[string]int)
for _, colour := range []string{"red", "blue", "red"} {
	votes[colour]++
}
// votes is map[string]int{"red": 2, "blue": 1}
```

> **Why is it written like this?** Go requires the map to exist before `++`
> writes to it, hence `make`. The names `votes` and `colour` are the author's
> choice.

### Ranging over only what you need

`for k, v := range m` gives keys and values. If you only need keys, drop the
value: `for k := range m`. If you only need values, use `_` for the key:
`for _, v := range m`.

> **Why is it written like this?** Go requires every declared variable to be
> used. `_` lets you skip the part you don't need.

### Returning several values

A function can return more than one value. The types go in parentheses, and
`return` lists the values in the same order:

```go
func MinMax(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

low, high := MinMax(7, 3) // low is 3, high is 7
```

## Your task

The catalog is a `map[string]string` from service name (the key) to the team
that owns it (the value):

```go
owners := map[string]string{
	"checkout": "payments",
	"invoices": "payments",
	"search":   "discovery",
}
```

Write the three functions in `exercise.go`. The tests compare maps with
`maps.Equal` and slices with `slices.Equal`, so a `nil` result and an empty
one count as the same.

### 1. Count services per team

**Goal:** `ServicesPerTeam(owners)` returns a map from each team to how many services it owns.

The keys of the result are **teams** (the values of `owners`). An empty or
`nil` catalog gives an empty map.

```go
ServicesPerTeam(owners)
// => map[string]int{"payments": 2, "discovery": 1}

ServicesPerTeam(nil)
// => map[string]int{} (empty)
```

**Think it through:**
- Which part of each `owners` entry do you count: the key or the value?
- Look at *Counting with a map* for how to count without checking first.

### 2. Find unowned services

**Goal:** `Unowned(services, owners)` returns the services from the list that nobody owns.

A service is unowned when it's **missing** from `owners`, or when its team is
the empty string `""`. Keep them in the same order as `services`. If every
service is owned, return an empty (or `nil`) slice. If `owners` is `nil`,
every service is unowned.

```go
owners := map[string]string{"checkout": "payments", "legacy-cron": ""}

Unowned([]string{"checkout", "legacy-cron", "metrics"}, owners)
// => []string{"legacy-cron", "metrics"}

Unowned([]string{"checkout"}, owners)
// => []string{} (empty)
```

**Think it through:**
- Loop over `services`, not over `owners`: that keeps the order.
- What does `owners[s]` give for a missing service? Can one check cover both cases?
- Grow the result with `append`, as in *Slices*.

### 3. Diff two catalogs

**Goal:** `Diff(before, after)` returns three counts: added, removed, changed, in that order.

- `added`: services in `after` that aren't in `before`.
- `removed`: services in `before` that aren't in `after`.
- `changed`: services in both, whose team is different.

Either catalog may be `nil`.

```go
before := map[string]string{"checkout": "payments", "search": "discovery", "legacy": "platform"}
after := map[string]string{"checkout": "payments", "search": "growth", "billing": "payments"}

Diff(before, after) // => 1, 1, 1
// added "billing", removed "legacy", changed "search"

Diff(before, before) // => 0, 0, 0
```

**Think it through:**
- Here a team of `""` is a real value, so you need `ok` to know if a service is there.
- One loop over `before` can find *removed* and *changed*. What finds *added*?
- The return order is fixed: added, removed, changed.

## Learn more

- [Go blog: Go maps in action](https://go.dev/blog/maps)
- [Go by Example: Maps](https://gobyexample.com/maps)
- [Go by Example: Multiple Return Values](https://gobyexample.com/multiple-return-values)
- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
