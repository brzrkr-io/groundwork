## Task 1: create the timer once

Move `time.After(window)` out of the `select`. Store it in a variable before
the `for` loop, and use that variable in the `case`:

```go
deadline := time.After(window)
for {
	select {
	// ...
	case <-deadline:
		return collected
	}
}
```

## Task 2: ticker plus stop

Create the ticker with `ticker := time.NewTicker(interval)` and add
`defer ticker.Stop()`. Then loop forever with a `select` that has two cases:
`<-ticker.C` and `<-stop`.

## Task 2: almost the answer

```go
ticker := time.NewTicker(interval)
defer ticker.Stop()
for {
	select {
	case <-ticker.C:
		beat()
		beats++
	case <-stop:
		return beats
	}
}
```
