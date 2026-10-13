## Task 1: look at the receiver

`Take` changes `l.Tokens`, but `(l RateLimiter)` hands it a copy. Which
receiver keeps the change?

## Task 2: the copy happens one step earlier

`TakeN` calls `l.Take()`. Even once `Take` has a pointer receiver, Go
rewrites the call as `(&l).Take()`, and inside `TakeN` the `l` is a copy of
the caller's limiter. Give `TakeN` a pointer receiver too, so `l` already is
a pointer to the real bucket.

## Task 3: loop over indexes

In `for _, lim := range limiters`, `lim` is a fresh copy of each element.
Loop with `for i := range limiters` and call the method on `limiters[i]`, the
element inside the slice.

## Almost the answer

Change the first line of `Take` and `TakeN` (keep their bodies):

```go
func (l *RateLimiter) Take() bool {
func (l *RateLimiter) TakeN(n int) int {
```

And loop by index in `RefillAll`:

```go
func RefillAll(limiters []RateLimiter, n int) {
	for i := range limiters {
		limiters[i].Refill(n)
	}
}
```
