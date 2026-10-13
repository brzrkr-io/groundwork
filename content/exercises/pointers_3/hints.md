## Task 1: Write WithReplicas

`spec.Replicas` is a `*int`, so you need a pointer to a new int that holds `n`. `spec` is already a copy, so set the field and return it.

### Closer

`new(n)` makes a new `int` holding `n` and gives you a pointer to it: `spec.Replicas = new(...)`.

### Answer

```go
spec.Replicas = new(n)
return spec
```

## Task 2: Write EffectiveReplicas

A `nil` pointer means "not set", and reading through it would crash. Check for that case first.

### Closer

```go
if spec.Replicas == nil {
	return DefaultReplicas
}
```

After that check it is safe to read the value with `*`.

### Answer

```go
if spec.Replicas == nil {
	return DefaultReplicas
}
return *spec.Replicas
```

## Task 3: Write Merge

Start from a copy of `base`, then go through `layers` in order. For each layer, only fields that aren't `nil` change the result.

### Closer

Use one `if` per field inside the loop, copying the pointed-to value:

```go
cfg := base
for _, o := range layers {
	if o.Image != nil {
		cfg.Image = *o.Image
	}
	// ...the same for Replicas and Debug
}
```

### Answer

```go
cfg := base
for _, o := range layers {
	if o.Image != nil {
		cfg.Image = *o.Image
	}
	if o.Replicas != nil {
		cfg.Replicas = *o.Replicas
	}
	if o.Debug != nil {
		cfg.Debug = *o.Debug
	}
}
return cfg
```
