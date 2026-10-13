## Task 1: the shape

Declare `var wg sync.WaitGroup` before the loop. In the loop:
`wg.Go(func() { deploy(region) })`. After the loop: `wg.Wait()`.

## Task 2: one slot per host

Keep `addrs[i] = resolve(host)`, but move it inside `wg.Go(func() { ... })`.
Call `wg.Wait()` after the loop and before `return addrs`.

## Task 3: let wg.Go do the counting

`wg` is already a pointer, so call its methods directly. Replace the three
lines with:

```go
wg.Go(func() {
	backup(db)
})
```
