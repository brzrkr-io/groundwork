## Task 1: Create the group and wait for it

`WithContext` needs a cancelable child of `ctx`; `context.WithCancel` gives
you one plus its cancel function. `Wait` waits for the WaitGroup, cancels,
and returns the saved error.

### Closer

In `WithContext`: `ctx, cancel := context.WithCancel(ctx)`, then return a
`&Group{cancel: cancel}` and the new `ctx`. In `Wait`: `g.wg.Wait()`, then
`g.cancel()`, then `return g.err`.

### Answer

```go
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Wait() error {
	g.wg.Wait()
	g.cancel()
	return g.err
}
```

Cancelling in `Wait` releases the context once every task is done.

## Task 2: Run tasks and keep the first error

The starter calls `f()` right away, so tasks run one after another. Run `f`
in a goroutine that `g.wg` tracks, and on an error let `g.once` decide
whether this is the first one.

### Closer

Use `g.wg.Go(func() { ... })`. Inside, `if err := f(); err != nil`, call
`g.once.Do(func() { ... })` with a function that sets `g.err = err` and
calls `g.cancel()`.

### Answer

```go
func (g *Group) Go(f func() error) {
	g.wg.Go(func() {
		if err := f(); err != nil {
			g.once.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	})
}
```

`once.Do` runs only for the first failure, so later `context.Canceled` errors are ignored.

## Task 3: Deploy every region, stop on the first failure

Create a group from the caller's `ctx`, start one task per region, and wait.
Each `deploy` call must get the **group's** context so a failure cancels the
others.

### Closer

`g, ctx := WithContext(ctx)` replaces `ctx` with the group's context, so
every `deploy(ctx, region)` after that line uses it. In the loop over
`regions`, call `g.Go(func() error { return deploy(ctx, region) })`, then
`return g.Wait()`.

### Answer

```go
g, ctx := WithContext(ctx)
for _, region := range regions {
	g.Go(func() error {
		return deploy(ctx, region)
	})
}
return g.Wait()
```

The group's context is a child of the caller's, so it is cancelled by either a failed region or the caller.
