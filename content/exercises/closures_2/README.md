# Functions as values: state that sticks

A closure is only useful if its state survives between calls. Here you fix a
counter that forgets everything, then build an ID generator and a rate
limiter for an API gateway.

## Learn

### Recap

A *closure* is a function literal that uses variables from the function
around it, as in *Functions as values*. Those variables live as long as the closure
does. Where you declare them decides how long they last:

```go
func NewTally() func() int {
	total := 0 // declared once, when NewTally runs: shared by every call
	return func() int {
		total++ // total from the outer function
		return total
	}
}
```

A variable declared *inside* the returned function is created fresh every
time that function runs, so it starts from zero on every call.

## Your task

### 1. Fix the error counter

`NewErrorCounter` should return a function that returns 1, 2, 3, ... on
successive calls, but it returns 1 every time. Move the count so it survives
between calls. Each counter counts on its own.

```go
count := NewErrorCounter()
count() // => 1
count() // => 2
count() // => 3
```

### 2. Build an ID generator

Change `NewIDGenerator` so it returns a function that returns `prefix`, a
`-`, and a number that starts at 1 and grows by one per call. Use
`strconv.Itoa` (you'll need to add the import) or `fmt.Sprintf`.

```go
next := NewIDGenerator("req")
next() // => "req-1"
next() // => "req-2"
NewIDGenerator("job")() // => "job-1"
```

### 3. Build a rate limiter

Change `NewRateLimiter` so it returns a function that returns `true` for the
first `limit` calls and `false` for every call after that. With a `limit` of
0, every call returns `false`.

```go
allow := NewRateLimiter(2)
allow() // => true
allow() // => true
allow() // => false
allow() // => false
```

## Learn more

- [A Tour of Go: Function closures](https://go.dev/tour/moretypes/25)
- [Go by Example: Closures](https://gobyexample.com/closures)
- [Go spec: Function literals](https://go.dev/ref/spec#Function_literals)
