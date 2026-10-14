# Pointer receivers: the leaky rate limiter

A rate limiter protects an API from being flooded. This one has three bugs,
and all three are the same kind: a change that lands on a copy instead of the
real limiter.

**New here:** pointer methods called on copies. (`for range n` and `min` are from *For loops: backoff and progress bars* and *For loops: plan a rollout*.)

## Learn

### Recap

A value receiver `(x T)` gives the method a copy, so field changes vanish
when it returns. A pointer receiver `(x *T)` points at the original, so
changes stick. Calling a pointer method on a variable, Go adds the `&` for
you: `v.M()` becomes `(&v).M()`.

### A pointer method called on a copy

The automatic `&` takes the address of *whatever variable you call it on*.
If that variable is itself a copy, the pointer method changes the copy.

This example uses a `Disk` whose `Write` method has a pointer receiver:

```go
// 1. Inside a value-receiver method: d is already a copy.
func (d Disk) WriteTwice(gb int) {
	d.Write(gb) // becomes (&d).Write(gb): changes the copy
	d.Write(gb) // the caller's disk never changes
}

// 2. In a range loop: disk is a copy of each element.
for _, disk := range disks {
	disk.Write(1) // becomes (&disk).Write(1): the slice is untouched
}
```

So a bug can sit in a method that *calls* the real changer, not just in the
changer itself.

### Reaching the real element of a slice

`disks[i]` is the element that lives inside the slice, not a copy. Calling a
pointer method on it changes the slice's element:

```go
disks[0].Write(1) // changes the first disk in the slice
```

> **Why is it written like this?** In `for i, v := range s`, Go requires
> that `v` is a copy. That is how the language works, not a style choice.
> `for i := range s` is the convention when you only need the index.

### Two handy shortcuts in the starter

- `for range n { ... }` runs the body `n` times, with no loop variable.
- `min(a, b)` is built in and returns the smaller value: `min(12, 10)` is
  `10`.

## Your task

Your API gateway gives each client a *token bucket*. The bucket holds up to
`Max` tokens. Each request takes one token, and a client with no tokens left
is turned away until the bucket is refilled. Find and fix the three bugs.

### 1. Fix Take

**Goal:** make `Take` really remove a token from the caller's limiter.

If the bucket has at least one token, `Take` removes one and returns `true`.
If it's empty, it returns `false` and changes nothing. The body is already
right; the tokens just never go down.

```go
l := RateLimiter{Tokens: 2, Max: 5}
l.Take() // => true, l.Tokens == 1
l.Take() // => true, l.Tokens == 0
l.Take() // => false, l.Tokens == 0
```

**Think it through:**
- Does `Take` change a field? What kind of receiver does that need?

### 2. Fix TakeN

**Goal:** make the tokens `TakeN(n)` takes come out of the caller's bucket.

`TakeN` calls `Take` up to `n` times, stops early when `Take` returns
`false`, and returns how many tokens it took. The count is already right;
the bucket just doesn't change.

```go
l := RateLimiter{Tokens: 5, Max: 5}
l.TakeN(3) // => 3, l.Tokens == 2
l.TakeN(3) // => 2, l.Tokens == 0
```

**Think it through:**
- After task 1, `Take` changes whatever it's called on. What is `TakeN`'s
  receiver: the original or a copy?
- Look at case 1 in *A pointer method called on a copy*.

### 3. Fix RefillAll

**Goal:** make `RefillAll(limiters, n)` refill every limiter inside the
slice, not copies of them.

`Refill(n)` already works: it adds `n` tokens but never goes past `Max`.
`RefillAll` calls it for each limiter, yet the slice doesn't change.

```go
limiters := []RateLimiter{
	{Tokens: 0, Max: 10},
	{Tokens: 8, Max: 10},
}
RefillAll(limiters, 4)
// limiters[0].Tokens == 4
// limiters[1].Tokens == 10 (8 + 4 is capped at Max)
```

**Think it through:**
- In the loop, is `lim` the element in the slice or a copy?
- What do you need from `range` to reach the real element?

## Learn more

- [A Tour of Go: Methods and pointer indirection](https://go.dev/tour/methods/6)
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
- [Token bucket (Wikipedia)](https://en.wikipedia.org/wiki/Token_bucket)
