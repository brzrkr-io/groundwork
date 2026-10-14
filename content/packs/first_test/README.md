# Your first test

*Practice: config validator* closed *Errors*: your code now reports what went
wrong. This module adds the other half: proving it works, with tests you
write yourself.

Until now Groundwork wrote the tests and you wrote the code. From here on you
write tests too. A test is a small program that checks your real program, so
you find a broken config parser on your laptop, not during a deploy.

**New here:** `_test.go` files, `testing.T`, `t.Errorf`, `t.Fatalf`, mutation grading.

## Learn

### Test files

Tests live next to the code, in files whose names end in `_test.go`. They are
in the same package, so they can call every function in it.

`go test` builds and runs them. `go build` ignores them, so tests never end
up in the program you ship.

```go
// file: exercise_test.go
package ports

import "testing"

func TestDouble(t *testing.T) {
	// checks go here
}
```

A *test function* must:

- have a name that starts with `Test`, then a capital letter (`TestDouble`),
- take exactly one parameter, `t *testing.T`,
- return nothing.

`go test` finds every function like this and runs it. A test passes unless it
reports a failure.

> **Why is it written like this?** Go requires the `_test.go` file ending,
> the `Test` prefix and the `t *testing.T` parameter; otherwise `go test`
> doesn't see your function. The name `t` is convention (every Go test uses
> it). The part after `Test` is your choice, but it usually names the
> function being tested.

### Got and want

Every check has three steps:

1. Call the function and store the result in `got`.
2. Compare it with the value you `want`.
3. If they differ, report a failure.

```go
got := Double(4)
want := 8
if got != want {
	t.Errorf("Double(4) = %d, want %d", got, want)
}
```

> **Why is it written like this?** `got` and `want` are just variable
> names, a community convention. The message format "call = got, want want"
> is also convention: it tells the reader what was called and what went
> wrong. Go itself only needs the `if` and a call that reports failure.

### Errorf or Fatalf?

`t` has two common ways to fail:

| Method | What happens |
|---|---|
| `t.Errorf(format, args...)` | Marks the test failed, prints the message, **keeps going**. |
| `t.Fatalf(format, args...)` | Marks the test failed, prints the message, **stops this test now**. |

Both format their message like `fmt.Sprintf` (`%d` for numbers, `%q` for
quoted text, `%v` for anything, such as an error).

Use `t.Errorf` by default, so one run shows every problem. Use `t.Fatalf`
when carrying on makes no sense, for example after an unexpected error: if
the call failed, the number it returned is meaningless.

```go
size, err := ParseSize("10GB")
if err != nil {
	t.Fatalf("ParseSize(%q) returned error: %v", "10GB", err)
}
if size != 10 {
	t.Errorf("ParseSize(%q) = %d, want %d", "10GB", size, 10)
}
```

If you call the function again later in the same test, `size` and `err`
already exist. Use `=` instead of `:=` for the second call (`:=` creates
new variables, `=` reuses them).

### Checking that bad input is rejected

Sometimes only the error matters. Then you fail when there is **no** error
(`err == nil`):

```go
if _, err := ParseSize("lots"); err == nil {
	t.Errorf("ParseSize(%q) returned no error, want an error", "lots")
}
```

> **Why is it written like this?** The `_` is Go's way to throw away a
> return value you don't need; Go rejects unused variables. Putting
> `_, err := ...;` before the condition is optional syntax: it keeps `err`
> inside this one `if`.

### Running tests yourself

Groundwork runs the tests whenever you save. You can also run them in a
terminal from the exercise folder:

```text
go test                     run every test in the package
go test -v                  also print each test's name and result
go test -run TestDouble     run only tests whose names match
```

### How Groundwork grades your test

A test that never fails is useless, so Groundwork checks that yours catches
bugs. `TestGrader` (in `grader_test.go`) does two things:

1. It runs your test against the real code. Your test must **pass**.
2. It runs your test against *mutants*: copies of the code with one
   deliberate bug each. Your test must **fail** on every mutant.

If a mutant slips through, `TestGrader` names the bug you missed, such as
"your tests didn't notice when ParsePort accepted port "0"".

The `parsePort` variable in `exercise.go` is how the grader swaps mutants in.
Don't change it or `grader_test.go`.

## Your task

The deploy tool reads service ports from config files with `ParsePort`. The
code in `exercise.go` is already correct. Write `TestParsePort` in
`exercise_test.go` so a broken `ParsePort` can never sneak back in.

`ParsePort` takes text, and returns a whole number from 1 to 65535 and a nil
error. Anything else gives an error.

```go
ParsePort("8080")  // => 8080, nil
ParsePort("0")     // => 0, error
```

### 1. Check valid ports

**Goal:** in `TestParsePort`, check that two valid ports come back as the right numbers with no error.

Check both:

- `ParsePort("8080")` returns `8080` and no error,
- `ParsePort("443")` returns `443` and no error.

Fail with `t.Fatalf` if there is an error, and with `t.Errorf` if the number
is wrong. This catches mutants that reject valid ports or always return the
same number.

**Think it through:**
- `ParsePort` returns two values. What do you name each one?
- Which check comes first: the error or the number?
- For the second port, do the variables already exist?

### 2. Check out-of-range ports

**Goal:** in the same `TestParsePort`, check that `ParsePort("0")` and `ParsePort("65536")` both return an error.

These are just outside the range 1–65535. That edge is where a bug like `<`
instead of `<=` hides. Report a failure with `t.Errorf` if either call gives
no error. The returned number doesn't matter here.

**Think it through:**
- Which return value do you care about, and which can you throw away?
- What does `err` look like when the function *wrongly* accepted the port?

The exercise is done when `TestParsePort` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
- [Package testing](https://pkg.go.dev/testing)
- [Add a test (Go tutorial)](https://go.dev/doc/tutorial/add-a-test)
