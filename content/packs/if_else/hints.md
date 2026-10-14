## Task 1: Classify disk usage

Go runs only the first branch whose condition is true. So check the biggest
number first, then the next one, and let everything left over be `"ok"`.

### Closer

Start with `if usedPercent >= 90 {`, then add `else if usedPercent >= 80`.
Keep the starter's `return "ok"` at the end for everything below 80.

### Answer

```go
if usedPercent >= 90 {
	return "critical"
} else if usedPercent >= 80 {
	return "warning"
}
return "ok"
```

## Task 2: Clamp the replica count

Handle the two special cases first, each in its own `if` that returns
straight away. Whatever gets past both checks is already fine.

### Closer

The first check looks like this:

```go
if requested < 1 {
	return 1
}
```

Add a second `if` for "above 10", and keep `return requested` at the end.

### Answer

```go
if requested < 1 {
	return 1
}
if requested > 10 {
	return 10
}
return requested
```

## Task 3: Give scaling advice

First guard against zero replicas, before any division. Then work out the
requests per replica inside the `if` itself, and check it.

### Closer

Return `"no replicas"` when `replicas < 1`. Then write the short statement:
`if per := totalRequests / replicas; per > 100 {`, and continue with
`else if per < 20` and `else`.

### Answer

```go
if replicas < 1 {
	return "no replicas"
}
if per := totalRequests / replicas; per > 100 {
	return "scale up"
} else if per < 20 {
	return "scale down"
} else {
	return "steady"
}
```

The guard comes first because dividing by zero would crash.
