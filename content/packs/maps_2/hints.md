## Task 1: Respect scaled-down services

`replicas[service]` gives `0` both when the service is listed as `0` and when
it isn't listed at all. You need a way to ask the map whether the key is
really there.

### Closer

The "comma ok" form gives you the value and a `bool`:
`n, ok := replicas[service]`. Return the default only when `ok` is `false`.

### Answer

```go
n, ok := replicas[service]
if !ok {
	return 2
}
return n
```

`ok` is `true` for a listed service even when its value is `0`.

## Task 2: Count only what was really removed

Only add one to `removed` when the key was in `env` before you delete it.
Don't compare the value with `""`: a key can be there with an empty value.

### Closer

Check for the key with comma ok, ignoring the value with `_`:
`if _, ok := env[k]; ok { ... }`. Put the `delete` and the `removed++` inside.

### Answer

```go
for _, k := range keys {
	if _, ok := env[k]; ok {
		delete(env, k)
		removed++
	}
}
```

A repeated key is already gone the second time, so `ok` is `false` and it isn't counted.

## Task 3: Merge without crashing

`var merged map[string]string` gives you a `nil` map, and `maps.Copy` writes
into it, so it panics. `merged` needs to be a real, empty map before anything
is copied in.

### Closer

Create the map with `make(map[string]string)`. Don't write
`merged := defaults`: that's the same map, so `Merge` would change `defaults`.

### Answer

```go
merged := make(map[string]string)
maps.Copy(merged, defaults)
maps.Copy(merged, overrides)
return merged
```

`make` gives a new empty map, so writing to it is safe and leaves both arguments untouched.
