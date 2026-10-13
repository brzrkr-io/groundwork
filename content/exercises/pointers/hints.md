## Task 1: write through the pointer

`count` is an address, not a number. To change the number it points to, put
`*` in front of it on the left of `=`:

```go
*count = 0
```

## Task 2: check for nil first

Start `Scale` with `if d == nil { return }`. After that, `d.Replicas = ...`
sets the field on the caller's deployment. Use an `if` to turn a negative
`replicas` into `0` before you assign it.

## Task 3: loop by index

Loop with `for i := range ds` so you can take the address of the real
element with `&ds[i]`. A pointer to the `range` variable (`&d`) points at a
copy, and the test will notice.

## Almost the answer

```go
func Scale(d *Deployment, replicas int) {
	if d == nil {
		return
	}
	if replicas < 0 {
		replicas = 0
	}
	d.Replicas = replicas
}

func FindDeployment(ds []Deployment, name string) *Deployment {
	for i := range ds {
		if ds[i].Name == name {
			return &ds[i]
		}
	}
	return nil
}
```
