# Sorting: the noisiest alerts

During an incident, nobody reads 300 alerts. They read the top five. Here you
turn a stream of firing alerts into a short "top N" summary.

**New here:** nothing; practice with maps, sorting and slicing.

## Learn

### Counting with a map

As in *Maps*, `m[k]++` adds one to the count for `k`. A missing key starts
at `0`, so you never need to check first.

```go
seen := make(map[string]int)
seen["disk"]++
seen["disk"]++
// seen == map[disk:2]
```

> **Why is it written like this?** Go requires `make` (or a `map[...]...{}`
> literal) before you write to a map; writing to a `nil` map panics. The name
> `seen` is your choice.

### Sorting with your own rule

`slices.Sorted(maps.Keys(m))` gives the keys in alphabetical order.
`slices.SortFunc(s, func(a, b T) int {...})` reorders `s` by your rule:
return a negative number if `a` comes first, positive if `b` does, `0` if
it's a tie. `cmp.Compare(x, y)` produces that number for you.

```go
// Shortest word first:
slices.SortFunc(words, func(a, b string) int {
	return cmp.Compare(len(a), len(b))
})
```

| You want | Compare |
|---|---|
| smallest first | `cmp.Compare(x of a, x of b)` |
| biggest first | swap them: `b` before `a` |

When the first comparison returns `0`, compare something else (such as the
name) so the order is the same on every run.

> **Why is it written like this?** Go requires the function to take two
> values of the slice's type and return an `int`. Map keys come out in random
> order, which is why a tie-break matters.

### Taking the first n

`s[:n]` keeps the first `n` elements. It panics if `n > len(s)`, so check
the length first.

### Formatting a line

`fmt.Sprintf("%s has %d", "disk", 2)` returns `"disk has 2"`: `%s` takes a
string, `%d` an integer.

## Your task

Your alerting system gives you one string per firing alert: the name of the
service that fired it. Add `cmp`, `fmt`, `maps` and `slices` to the imports
of `exercise.go` as you need them.

### 1. Count alerts per service

**Goal:** make `CountAlerts` return a map from each service name to how many
times it appears in `firing`.

```go
CountAlerts([]string{"api", "db", "api", "api", "queue"})
// => map[api:3 db:1 queue:1]
```

An empty slice gives an empty map (not `nil`).

**Think it through:**
- The function returns `nil` now. Where do you create the map to return?
- What does the loop variable hold on each pass through `firing`?

### 2. Report the top N

**Goal:** make `TopAlerts` return up to `n` lines like `"api: 12"`, busiest
service first.

- Each line is: service name, then the fixed text `": "`, then the count.
- The service with the **most** alerts comes first.
- Services with the same count are in alphabetical order.
- Return at most `n` lines. Fewer than `n` services? Return all of them.
- `n` is `0` or negative, or `counts` is empty: return an empty slice.

```go
counts := map[string]int{"api": 12, "db": 3, "queue": 7, "cache": 3}

TopAlerts(counts, 3)
// => ["api: 12" "queue: 7" "cache: 3"]

TopAlerts(counts, 10)
// => ["api: 12" "queue: 7" "cache: 3" "db: 3"]
```

**Think it through:**
- Sort the *names*, but compare by the count each name has in `counts`.
- Which comparison goes first, and which one only runs on a tie?
- When is it safe to cut the slice down to `n`?

## Learn more

- [Package slices: Sorted and SortFunc](https://pkg.go.dev/slices)
- [Package cmp: Compare](https://pkg.go.dev/cmp#Compare)
- [Go by Example: Maps](https://gobyexample.com/maps)
