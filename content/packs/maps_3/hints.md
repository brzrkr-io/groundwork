## Task 1: Count services per team

Make a new map from team to count, then loop over `owners`. Each service adds
one to its team's count.

### Closer

Create the map with `make(map[string]int)`. You only need the team, so ignore
the service name with `_`: `for _, team := range owners`. Inside, `counts[team]++`
works even the first time, because a missing count reads as `0`.

### Answer

```go
counts := make(map[string]int)
for _, team := range owners {
	counts[team]++
}
return counts
```

Ranging over a `nil` map runs zero times, so `ServicesPerTeam(nil)` returns an empty map.

## Task 2: Find unowned services

Loop over `services`, not over the map, so the result keeps their order. Think
about what `owners[s]` gives you when `s` is missing.

### Closer

A missing service reads as `""`, so `owners[s] == ""` covers both "missing"
and "empty team". Collect matches with `unowned = append(unowned, s)` into a
`var unowned []string`.

### Answer

```go
var unowned []string
for _, s := range services {
	if owners[s] == "" {
		unowned = append(unowned, s)
	}
}
return unowned
```

One check catches both cases because a missing key gives the zero value `""`.

## Task 3: Diff two catalogs

Use two loops. One over `before` finds removed and changed services; one over
`after` finds added services.

### Closer

In the `before` loop, look the service up in `after` with comma ok:
`newTeam, ok := after[service]`. Not there means `removed++`; there with a
different team means `changed++`. In the `after` loop, use
`if _, ok := before[service]; !ok` to count `added++`.

### Answer

```go
added, removed, changed := 0, 0, 0
for service, oldTeam := range before {
	newTeam, ok := after[service]
	if !ok {
		removed++
	} else if newTeam != oldTeam {
		changed++
	}
}
for service := range after {
	if _, ok := before[service]; !ok {
		added++
	}
}
return added, removed, changed
```

Reading a `nil` map works like an empty one, so either catalog may be `nil`.
