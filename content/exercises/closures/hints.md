## Task 1: Build a threshold checker

The starter already returns a function literal. Its body just needs the right
comparison. The inner function can use `limit` directly.

### Closer

"At or above" is the `>=` operator. Compare the function's own parameter
`value` with `limit` and return the result.

### Answer

```go
return func(value int) bool {
	return value >= limit
}
```

## Task 2: Build a request counter

The count must survive between calls, so it can't live inside the returned
function. Where else can you declare it?

### Closer

Declare `count := 0` in `NewRequestCounter`, *before* the `return`. Inside the
returned function, add 1 with `count++`, then return `count`.

### Answer

```go
count := 0
return func() int {
	count++
	return count
}
```

## Task 3: Retry an operation

`try` is a function you can call like any other. Call it with attempt numbers
1, 2, 3, ... and stop as soon as one works.

### Closer

Use a `for` loop: `for attempt := 1; attempt <= attempts; attempt++`. Inside,
`if try(attempt)` is true, return `attempt` straight away. After the loop,
return `0`.

### Answer

```go
for attempt := 1; attempt <= attempts; attempt++ {
	if try(attempt) {
		return attempt
	}
}
return 0
```
