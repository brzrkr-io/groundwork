## Task 1: Write MemoryStore.Get

A map lookup can already tell you two things: the value, and whether the key was there. That's exactly what `Get` returns.

### Closer

Use the two-value form: `v, ok := m.Values[key]`. Reading from a nil map is fine; it just finds nothing.

### Answer

```go
v, ok := m.Values[key]
return v, ok
```

`ok` is `true` even when the stored value is `""`, so empty values count as found.

## Task 2: Write Resolve

Make an empty result map, then look up each key in `s`. Stop at the first missing key. Remember to add `"fmt"` to the imports.

### Closer

Create the map with `make(map[string]string)`, loop with `for _, key := range keys`, and call `s.Get(key)`. If `ok` is `false`, `return nil, fmt.Errorf("missing config key %q", key)`.

### Answer

```go
config := make(map[string]string)
for _, key := range keys {
	v, ok := s.Get(key)
	if !ok {
		return nil, fmt.Errorf("missing config key %q", key)
	}
	config[key] = v
}
return config, nil
```

`Resolve` only knows `s` is a `Store`, so it works with any store, real or fake.

## Task 3: Make Layered a Store

`Layered` needs a `Get` method with exactly the signature from `Store`. Inside, ask each layer in order and return as soon as one has the key.

### Closer

Each `layer` in `l.Layers` is a `Store`, so you can call `layer.Get(key)`:

```go
func (l Layered) Get(key string) (string, bool) {
	for _, layer := range l.Layers {
		// ask layer, return if found
	}
	return "", false
}
```

### Answer

```go
// Get returns the value from the first layer that has key.
func (l Layered) Get(key string) (string, bool) {
	for _, layer := range l.Layers {
		if v, ok := layer.Get(key); ok {
			return v, true
		}
	}
	return "", false
}
```

Earlier layers are asked first, so overrides win over defaults.
