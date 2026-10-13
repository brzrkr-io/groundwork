## Task 1: order matters

`wg.Wait()` is what makes the goroutines' writes safe to read. Move it above
the loop that adds up `total`.

## Task 2: a slice of errors

Make `errs := make([]error, len(images))` and use `for i, image := range
images`. Each goroutine stores `errs[i] = push(image)`. After `wg.Wait()`,
loop over `errs` and return the first one that is not `nil`. Return `nil` if
there is none.

## Task 3: once.Do

Replace the `if` with:

```go
c.once.Do(func() {
	c.token = c.fetch()
})
return c.token
```

Every caller waits for `Do` to finish, so reading `c.token` afterwards is
safe.
