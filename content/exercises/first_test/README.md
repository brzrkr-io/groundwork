# Your first test

Until now Gopherlings wrote the tests and you wrote the code. From here on you
write tests too. A test is a small program that checks your real program, so
you find out a config parser is broken on your laptop instead of during a
deploy.

## Learn

### Test files

Tests live next to the code, in files whose names end in `_test.go`. They
belong to the same package, so they can call every function in it. The
`go test` command builds and runs them. A normal `go build` ignores them, so
tests never end up in the program you ship.

```go
// file: exercise_test.go
package ports

import "testing"

func TestParsePort(t *testing.T) {
	// checks go here
}
```

A *test function* must:

- start with `Test`, followed by a capital letter (`TestParsePort`),
- take exactly one parameter, `t *testing.T`, from the `testing` package,
- return nothing.

`go test` finds every function like this and runs it. A test passes unless it
reports a failure.

### Reporting failures

The usual shape is *got* and *want*: call the function, then compare what you
got with what you want.

```go
got := Double(4)
want := 8
if got != want {
	t.Errorf("Double(4) = %d, want %d", got, want)
}
```

`t` gives you two common ways to fail:

| Method | What happens |
|---|---|
| `t.Errorf(format, args...)` | Marks the test failed, prints the message, and **keeps going**. |
| `t.Fatalf(format, args...)` | Marks the test failed, prints the message, and **stops this test** right away. |

Both format their message like `fmt.Sprintf`. Use `t.Errorf` by default, so
one run shows you every problem. Use `t.Fatalf` when carrying on makes no
sense. A common case is an unexpected error: if `ParsePort` failed, the port
it returned is meaningless, so stop there.

```go
port, err := ParsePort("22")
if err != nil {
	t.Fatalf("ParsePort(%q) returned error: %v", "22", err)
}
if port != 22 {
	t.Errorf("ParsePort(%q) = %d, want %d", "22", port, 22)
}
```

To check that bad input is rejected, look for a missing error:

```go
if _, err := ParsePort("ssh"); err == nil {
	t.Errorf("ParsePort(%q) returned no error, want an error", "ssh")
}
```

The `_` throws away the port, because only the error matters here.

### Running tests yourself

Gopherlings runs the tests whenever you save. You can also run them in a
terminal from the exercise folder:

```text
go test                       run every test in the package
go test -v                    also print each test's name and result
go test -run TestParsePort    run only tests whose names match
```

### How Gopherlings grades your test

A test that never fails is useless, so Gopherlings checks that yours catches
bugs. The file `grader_test.go` holds `TestGrader`. It re-runs your test
against the real `ParsePort`, where it must pass. Then it runs it again
against several *mutants*: copies of `ParsePort` with one deliberate bug each.
Your test must fail against every mutant. If one slips through, `TestGrader`
names the bug your test missed, for example "your tests didn't notice when
ParsePort accepted port "0"". The `parsePort` variable in `exercise.go` is how
the grader swaps the mutants in. You don't need to touch it or
`grader_test.go`.

## Your task

The deploy tool reads service ports from config files with `ParsePort`. The
code in `exercise.go` is already correct. Write `TestParsePort` in
`exercise_test.go` so that a broken `ParsePort` can never sneak back in.

`ParsePort` accepts a whole number from 1 to 65535 and returns it. Anything
else gives an error.

```go
ParsePort("8080")  // => 8080, nil
ParsePort("0")     // => 0, error
```

### 1. Check valid ports

In `TestParsePort`, check that:

- `ParsePort("8080")` returns `8080` and no error,
- `ParsePort("443")` returns `443` and no error.

Fail with `t.Fatalf` if there is an error and `t.Errorf` if the number is
wrong. This catches mutants that reject valid ports or return the wrong
number.

### 2. Check out-of-range ports

Check that `ParsePort("0")` and `ParsePort("65536")` both return an error.
These are just outside the allowed range, the places where a bug like `<`
instead of `<=` hides.

The exercise is done when `TestParsePort` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
- [Package testing](https://pkg.go.dev/testing)
- [Add a test (Go tutorial)](https://go.dev/doc/tutorial/add-a-test)
