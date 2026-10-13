## Task 1: two small functions

```go
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}
```

`Wait` is three lines: `g.wg.Wait()`, `g.cancel()`, `return g.err`.

## Task 2: once.Do inside the goroutine

Start the goroutine with `g.wg.Go(func() { ... })`. Inside it, call `f()`.
If it returns an error, call `g.once.Do` with a function that sets `g.err`
and calls `g.cancel()`.

```go
g.wg.Go(func() {
	if err := f(); err != nil {
		g.once.Do(func() {
			g.err = err
			g.cancel()
		})
	}
})
```

## Task 3: shadow ctx on purpose

`g, ctx := WithContext(ctx)` replaces `ctx` with the group's context. Every
`deploy(ctx, region)` after that line gets the group's context. Start one
`g.Go` per region and `return g.Wait()`.
