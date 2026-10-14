## Task 1: Read after waiting

The goroutines are fine. Look at when the loop that adds up `total` runs
compared with `wg.Wait()`.

### Closer

`wg.Wait()` is what makes the goroutines' writes to `sizes` safe to read.
Nothing may read `sizes` before it.

### Answer

```go
wg.Wait()
for _, s := range sizes {
	total += s
}
return sizes, total
```

After `Wait`, every goroutine has finished writing, so reading is safe.

## Task 2: Stop sharing err

Every goroutine writes the same `err`. Give each goroutine its own slot to
write, like `sizes[i]` in task 1, then look at the slots after waiting.

### Closer

Make `errs := make([]error, len(images))` and range with
`for i, image := range images`. Each goroutine stores `errs[i] = push(image)`.
After `wg.Wait()`, loop over `errs` in order.

### Answer

```go
errs := make([]error, len(images))
var wg sync.WaitGroup
for i, image := range images {
	wg.Go(func() {
		errs[i] = push(image)
	})
}
wg.Wait()
for _, err := range errs {
	if err != nil {
		return err
	}
}
return nil
```

Each goroutine owns one slot, and looping in order picks the first image's error.

## Task 3: Fetch the token once

Checking `c.token == ""` and then setting it lets two goroutines both see
`""`. The struct already has a `sync.Once` field for exactly this.

### Closer

Replace the `if` with `c.once.Do(func() { ... })`, setting `c.token` inside
the function, then return `c.token`.

### Answer

```go
func (c *TokenCache) Token() string {
	c.once.Do(func() {
		c.token = c.fetch()
	})
	return c.token
}
```

Every caller waits for `Do` to finish, so reading `c.token` afterwards is safe.
