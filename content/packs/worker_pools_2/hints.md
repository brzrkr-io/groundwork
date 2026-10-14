## Task 1: Fix the hang

The collector's `for p := range results` only ends when `results` is closed.
Something has to close it after every worker is done, but without blocking the
collector loop that comes after it.

### Closer

Start one more goroutine after the workers. It calls `wg.Wait()`, then
`close(results)`. It must be a goroutine: calling `wg.Wait()` directly would
block before the collector starts, and the workers would be stuck sending.

### Answer

```go
go func() {
	wg.Wait()
	close(results)
}()
```

The closer runs alongside the collector, so the workers can keep sending until they finish.

## Task 2: Keep the input order

`append` adds results in the order they arrive, which is finish order. Each
`pulled` value already carries `index`, the image's position in `images`.

### Closer

Make the slice full-size up front with `make([]Result, len(images))`, then
write into a slot instead of appending: `out[...] = p.result`.

### Answer

```go
out := make([]Result, len(images))
for p := range results {
	out[p.index] = p.result
}
return out
```

Each result lands in its image's slot, whatever order the pulls finish in.
