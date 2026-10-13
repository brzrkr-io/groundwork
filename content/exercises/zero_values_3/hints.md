## Task 1: one if per layer

Check `requestSeconds` first. If it isn't `0`, return it straight away. Then
do the same for `serviceSeconds` and `globalSeconds`.

## Task 1: the shape

```go
if requestSeconds != 0 {
	return requestSeconds
}
// ...two more checks...
return 30
```

## Task 2: defaults, then format

Overwrite each parameter when it is zero, then build the string once:

```go
return fmt.Sprintf("GET %s on :%d every %ds", path, port, periodSeconds)
```
