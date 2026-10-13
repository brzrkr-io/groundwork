## Task 1: the map already answers

Looking up a map key with two variables gives you the value and whether the
key was there: `v, ok := m.Values[key]`. That is exactly what `Get` returns.
Reading from a nil map is fine; it just finds nothing.

## Task 2: loop, look up, stop early

Make the result map with `make(map[string]string)`. Loop over `keys`, call
`s.Get(key)`, and if `ok` is `false` return `nil` and the error straight
away. Otherwise store the value in the map. Return the map and `nil` after the
loop.

## Task 3: the method signature

`Layered` needs a method with exactly the signature from `Store`:

```go
// Get returns the value from the first layer that has key.
func (l Layered) Get(key string) (string, bool) {
	for _, layer := range l.Layers {
		// ask layer, and return as soon as one has it
	}
	return "", false
}
```

## Task 3: inside the loop

Each `layer` is a `Store`, so you can call `layer.Get(key)`:

```go
if v, ok := layer.Get(key); ok {
	return v, true
}
```
