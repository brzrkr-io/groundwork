## Task 1: count the values

Make a `map[string]int`, then loop over `owners`. You only need the team, so
ignore the service name with `_`, and do `counts[team]++`, as in *Maps*.

## Task 2: loop over the slice, not the map

Loop over `services` so the result keeps their order, and `append` each
unowned one to a result slice. A missing service reads as `""` from the map,
so `owners[s] == ""` covers both "missing" and "empty team".

## Task 3: two loops

Loop over `before`. For each service, look it up in `after` with comma ok:
not there means `removed++`, there with a different team means `changed++`.
Then loop over `after` and count `added++` for every service that isn't in
`before`.

```go
for service, oldTeam := range before {
	newTeam, ok := after[service]
	if !ok {
		removed++
	} else if newTeam != oldTeam {
		changed++
	}
}
```
