## Task 1: Fix the error counter

`count := 0` sits inside the returned function, so it starts again at 0 every
time that function runs. Where should it live so it is created only once?

### Closer

Move the `count := 0` line up, out of the function literal, so it sits above
`return func() int {`. Keep `count++` and `return count` inside.

### Answer

```go
count := 0
return func() int {
	count++
	return count
}
```

## Task 2: Build an ID generator

Same idea as task 1: keep a number outside the returned function and add 1 on
every call. Then join `prefix`, a dash and the number into one string.

### Closer

Declare `n := 0` before the `return`, and do `n++` inside. Fixed text goes in
quotes (`"-"`); `prefix` doesn't. `strconv.Itoa(n)` turns the number into text,
so join with `+`: `prefix + "-" + ...`. Add `strconv` to the imports.

### Answer

```go
n := 0
return func() string {
	n++
	return prefix + "-" + strconv.Itoa(n)
}
```

## Task 3: Build a rate limiter

Count how many calls have happened so far. A call is allowed while that count
is still within `limit`.

### Closer

Declare `calls := 0` outside the returned function. Inside, do `calls++`, then
return the result of comparing `calls` with `limit` using `<=`.

### Answer

```go
calls := 0
return func() bool {
	calls++
	return calls <= limit
}
```
