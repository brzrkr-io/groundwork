## Task 1: slots plus Join

```go
errs := make([]error, len(tasks))
var wg sync.WaitGroup
for i, task := range tasks {
	wg.Go(func() {
		errs[i] = task()
	})
}
wg.Wait()
return errors.Join(errs...)
```

## Task 2: take a slot before starting

Copy task 1, and add `sem := make(chan struct{}, limit)`. In the loop, before
`wg.Go`, send `sem <- struct{}{}`. Inside the goroutine, start with
`defer func() { <-sem }()`.

## Task 3: build the tasks, then reuse RunLimited

```go
var tasks []func() error
for _, host := range hosts {
	tasks = append(tasks, func() error {
		if err := check(host); err != nil {
			return fmt.Errorf("%s: %w", host, err)
		}
		return nil
	})
}
return RunLimited(tasks, limit)
```

Return `nil` (not a wrapped `nil`) for healthy hosts, so `errors.Join` skips
them.
