## Task 1: one struct literal

Replace `Deployment{}` with a struct literal that sets all five fields with
`Field: value` pairs. `Limits` is a struct too, so its value is another
literal: `Resources{CPUMillis: 250, MemoryMB: 128}`. For `Env`, use
`map[string]string{}`.

## Task 2: check, then assign

`d` is your own copy, so you can change it and return it. Reach the nested
field with two dots:

```go
if d.Limits.MemoryMB == 0 {
	d.Limits.MemoryMB = 128
}
```

Do the same for `d.Replicas`.

## Task 3: loop and compare names

```go
for _, d := range ds {
	if d.Name == name {
		return d, true
	}
}
return Deployment{}, false
```
