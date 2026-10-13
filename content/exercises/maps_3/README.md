# Maps: diff a service catalog

Every platform team ends up with a *service catalog*: which team owns which
service. In this exercise you answer the questions people ask of it during a
reorg, using maps from end to end.

## Learn

### Recap

Read with `m[k]` (zero value if missing) or `v, ok := m[k]` (to know whether
it's there). Write with `m[k] = v` after creating the map with `make` or a
literal. Count with `m[k]++`. Loop with `for k, v := range m`, in random
order. Reading a `nil` map is fine and behaves like an empty one.

## Your task

The catalog is a `map[string]string` from service name to the team that owns
it:

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

`ServicesPerTeam` returns a map from each team to how many services it owns.

```go
ServicesPerTeam(owners)
// => map[string]int{"payments": 2, "discovery": 1}

ServicesPerTeam(nil)
// => map[string]int{} (empty)
```

### 2. Find unowned services

`Unowned` gets a list of running services and returns the ones nobody owns,
in the same order as `services`. A service is unowned when it's missing from
`owners`, or when its team is the empty string `""`.

```go
owners := map[string]string{"checkout": "payments", "legacy-cron": ""}

Unowned([]string{"checkout", "legacy-cron", "metrics"}, owners)
// => []string{"legacy-cron", "metrics"}

Unowned([]string{"checkout"}, owners)
// => []string{} (empty)
```

### 3. Diff two catalogs

`Diff` compares the catalog `before` a reorg with the one `after` it, and
returns three counts:

- `added`: services in `after` that aren't in `before`,
- `removed`: services in `before` that aren't in `after`,
- `changed`: services in both, whose team is different.

```go
before := map[string]string{"checkout": "payments", "search": "discovery", "legacy": "platform"}
after := map[string]string{"checkout": "payments", "search": "growth", "billing": "payments"}

Diff(before, after) // => 1, 1, 1
// added "billing", removed "legacy", changed "search"

Diff(before, before) // => 0, 0, 0
```

Either catalog may be `nil`.

## Learn more

- [Go blog: Go maps in action](https://go.dev/blog/maps)
- [Go by Example: Maps](https://gobyexample.com/maps)
- [Go by Example: Range over Built-in Types](https://gobyexample.com/range-over-built-in-types)
