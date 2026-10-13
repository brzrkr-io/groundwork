# Sorting: the noisiest alerts

During an incident, nobody reads 300 alerts. They read the top five. In this
exercise you turn a stream of firing alerts into a short "top N" summary for
the incident channel.

## Learn

### Recap

Count things with a map and `m[k]++`, as in *Maps*. Get the keys with
`slices.Sorted(maps.Keys(m))`, then reorder them with `slices.SortFunc` and
`cmp.Compare`: swap the arguments for "biggest first", and tie-break on the
name so the output never changes between runs. Take the first `n` elements of
a slice with `s[:n]`, which panics if `n` is bigger than `len(s)`. Build each
line with `fmt.Sprintf`.

## Your task

Your alerting system gives you one string per firing alert: the name of the
service that fired it. You write the code that summarises them.

The solution needs the `cmp`, `fmt`, `maps` and `slices` packages. Add them to
the imports at the top of `exercise.go` as you go.

### 1. Count alerts per service

Change `CountAlerts` to return a map from service name to the number of times
that name appears in `firing`.

```go
CountAlerts([]string{"api", "db", "api", "api", "queue"})
// => map[api:3 db:1 queue:1]
```

An empty slice gives an empty map.

### 2. Report the top N

Change `TopAlerts` to return one line per service, formatted as
`"<service>: <count>"`:

- The service with the **most** alerts comes first.
- Services with the same count are in alphabetical order.
- Return at most `n` lines. If there are fewer than `n` services, return all of
  them.
- If `n` is `0` or negative, or `counts` is empty, return an empty slice.

```go
counts := map[string]int{"api": 12, "db": 3, "queue": 7, "cache": 3}

TopAlerts(counts, 3)
// => ["api: 12" "queue: 7" "cache: 3"]

TopAlerts(counts, 10)
// => ["api: 12" "queue: 7" "cache: 3" "db: 3"]
```

## Learn more

- [Package slices: Sorted and SortFunc](https://pkg.go.dev/slices)
- [Package cmp: Compare](https://pkg.go.dev/cmp#Compare)
- [Go by Example: Maps](https://gobyexample.com/maps)
