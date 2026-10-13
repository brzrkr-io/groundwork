## Task 1: return a function literal

The function you return can use `limit` directly:

```go
return func(value int) bool {
	return value >= limit
}
```

## Task 2: the count lives outside

Declare `count := 0` in `NewRequestCounter`, *before* the `return`. The
returned function adds 1 to it and returns it.

## Task 3: call try in a loop

Loop `attempt` from 1 to `attempts`, and return early on success:

```go
for attempt := 1; attempt <= attempts; attempt++ {
	if try(attempt) {
		return attempt
	}
}
return 0
```
