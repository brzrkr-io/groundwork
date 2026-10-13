## Task 1: new with a value

`new(n)` makes a new `int` holding `n` and gives you a pointer to it. Store
it in the copy and return the copy:

```go
spec.Replicas = new(n)
return spec
```

## Task 2: check for nil, then read

```go
if spec.Replicas == nil {
	return DefaultReplicas
}
```

After that check it is safe to return `*spec.Replicas`.

## Task 3: one if per field

Copy `base` into a result variable, then loop over `layers`. Inside the
loop, check each pointer field of the layer on its own: if it isn't `nil`,
copy the value it points to into the matching field of the result.

## Almost the answer

```go
func Merge(base Config, layers ...Overrides) Config {
	cfg := base
	for _, o := range layers {
		if o.Image != nil {
			cfg.Image = *o.Image
		}
		// ...the same for Replicas and Debug
	}
	return cfg
}
```
