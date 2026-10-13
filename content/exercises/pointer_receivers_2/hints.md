## Task 1: Fix Take

`Take` changes `l.Tokens`, but look at its receiver: `(l RateLimiter)` hands
it a copy. Which kind of receiver keeps the change?

### Closer

Add a `*` to the receiver type: `(l *RateLimiter)`. The body stays the same.

### Answer

```go
func (l *RateLimiter) Take() bool {
```

## Task 2: Fix TakeN

`TakeN` calls `l.Take()`, and `Take` now has a pointer receiver. But what is
`l` inside `TakeN`? It's still a copy of the caller's limiter.

### Closer

Go rewrites `l.Take()` as `(&l).Take()`, so it points at the copy. Give
`TakeN` a pointer receiver too, so `l` already is a pointer to the real
bucket.

### Answer

```go
func (l *RateLimiter) TakeN(n int) int {
```

Keep the body as it is; now every `Take` lands on the caller's bucket.

## Task 3: Fix RefillAll

In `for _, lim := range limiters`, `lim` is a fresh copy of each element, so
`lim.Refill(n)` refills the copy.

### Closer

Loop over indexes with `for i := range limiters` and call the method on
`limiters[i]`, the element that lives inside the slice.

### Answer

```go
for i := range limiters {
	limiters[i].Refill(n)
}
```
