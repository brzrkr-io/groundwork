# Pointer receivers: the leaky rate limiter

A rate limiter protects an API from being flooded. This one has three bugs,
and all three are the same kind: a change that lands on a copy instead of the
real limiter.

## Learn

### Recap

A value receiver `(l RateLimiter)` gives the method a copy, so field changes
vanish when it returns. A pointer receiver `(l *RateLimiter)` points at the
original, so changes stick. When you call a pointer method on a variable, Go
adds the `&` for you: `l.Take()` becomes `(&l).Take()`. As in
*Pointer receivers*, once one method needs a pointer receiver, give them all
one.

### A pointer method called on a copy

The automatic `&` takes the address of *whatever variable you call it on*. If
that variable is itself a copy, the pointer method dutifully changes the copy.
Two common places this happens:

```go
// 1. Inside a value-receiver method: l is already a copy.
func (l RateLimiter) TakeTwo() {
	l.Take() // becomes (&l).Take(): changes the copy
	l.Take() // the caller's limiter never changes
}

// 2. In a range loop: lim is a copy of each element.
for _, lim := range limiters {
	lim.Take() // becomes (&lim).Take(): the slice is untouched
}
```

To change the elements of a slice, use the index, so the method gets a pointer
to the element that lives inside the slice:

```go
for i := range limiters {
	limiters[i].Take() // changes limiters[i] itself
}
```

### Two handy shortcuts in the starter

- `for range n { ... }` runs the loop body `n` times. You don't need a loop
  variable when you only want to repeat something.
- `min(a, b)` is a built-in function that returns the smaller of its
  arguments. `min(12, 10)` is `10`.

## Your task

Your API gateway gives each client a *token bucket*. The bucket holds up to
`Max` tokens. Each request takes one token, and a client with no tokens left
is turned away until the bucket is refilled. Find and fix the three bugs.

### 1. Fix Take

`Take` handles one request. If the bucket has at least one token, it removes
one token and returns `true`. If it's empty, it returns `false` and changes
nothing. The logic is right, but the tokens never go down.

```go
l := RateLimiter{Tokens: 2, Max: 5}
l.Take() // => true, l.Tokens == 1
l.Take() // => true, l.Tokens == 0
l.Take() // => false, l.Tokens == 0
```

### 2. Fix TakeN

`TakeN(n)` handles a batch of `n` requests by calling `Take` up to `n` times.
It stops early when `Take` returns `false`, and returns how many tokens it
actually took. `TakeN` reports the right number, but the bucket still doesn't
change.

```go
l := RateLimiter{Tokens: 5, Max: 5}
l.TakeN(3) // => 3, l.Tokens == 2
l.TakeN(3) // => 2, l.Tokens == 0
```

### 3. Fix RefillAll

`Refill(n)` adds `n` tokens, but never fills a bucket past `Max`. It already
works. `RefillAll(limiters, n)` should call `Refill(n)` on every limiter in
the slice, yet the limiters in the slice don't change. Make the change land
on the slice's elements.

```go
limiters := []RateLimiter{
	{Tokens: 0, Max: 10},
	{Tokens: 8, Max: 10},
}
RefillAll(limiters, 4)
// limiters[0].Tokens == 4
// limiters[1].Tokens == 10 (8 + 4 is capped at Max)
```

## Learn more

- [A Tour of Go: Methods and pointer indirection](https://go.dev/tour/methods/6)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
- [Token bucket (Wikipedia)](https://en.wikipedia.org/wiki/Token_bucket)
