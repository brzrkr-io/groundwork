## Task 1: comma ok

`n == 0` is true both for "listed as 0" and "not listed". Ask the map which
one it is with `n, ok := replicas[service]`, and return the default only when
`ok` is `false`.

## Task 2: look before you count

Inside the loop, check whether the key is in `env` before deleting it, and
only add one to `removed` when it was. Don't compare the value with `""`: a
key can be there with an empty value.

```go
_, ok := env[k]
if ok {
	delete(env, k)
	removed++
}
```

## Task 3: make the map

`var merged map[string]string` gives you a `nil` map, and `maps.Copy` writes
into it, so it panics. Create a real, empty map instead:

```go
merged := make(map[string]string)
```

Don't write `merged := defaults`: that's the same map, so `Merge` would change
`defaults`.
