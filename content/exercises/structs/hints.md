## Task 1: Build a deployment with defaults

Replace the empty `Deployment{}` with a struct literal that sets all five fields. `Limits` is itself a struct, so its value is another struct literal.

### Closer

Use `Field: value` pairs, one per line, each ending with a comma. The nested field looks like `Limits: Resources{CPUMillis: 250, MemoryMB: 128}`, and an empty map is `map[string]string{}`.

### Answer

```go
return Deployment{
	Name:     name,
	Image:    image,
	Replicas: 1,
	Env:      map[string]string{},
	Limits:   Resources{CPUMillis: 250, MemoryMB: 128},
}
```

## Task 2: Fill in missing settings

`d` is your own copy, so you can change its fields and return it. Check each field for `0` before you set it.

### Closer

Reach the nested field with two dots, `d.Limits.MemoryMB`:

```go
if d.Limits.MemoryMB == 0 {
	d.Limits.MemoryMB = 128
}
```

Do the same for `d.Replicas`.

### Answer

```go
if d.Replicas == 0 {
	d.Replicas = 1
}
if d.Limits.MemoryMB == 0 {
	d.Limits.MemoryMB = 128
}
return d
```

## Task 3: Find a deployment by name

Loop over `ds` and compare each deployment's `Name` with `name`. Return as soon as you find a match; the existing last line handles "not found".

### Closer

`for _, d := range ds` gives you each deployment in turn. Inside, `if d.Name == name` tells you it's the one, so return it together with `true`.

### Answer

```go
for _, d := range ds {
	if d.Name == name {
		return d, true
	}
}
return Deployment{}, false
```
