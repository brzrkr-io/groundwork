# Your first test: fix a broken test

Tests have bugs too. A test that can never fail is worse than no test at
all, because it makes you feel safe when you aren't.

## Learn

### Recap

A test is a `func TestXxx(t *testing.T)` in a `_test.go` file. It calls your
code, compares *got* with *want*, and reports a mismatch with `t.Errorf` (keep
going) or `t.Fatalf` (stop this test). As in *Your first test*, `TestGrader`
re-runs your test against broken copies of the code, called mutants. Your test
must pass on the real code and fail on every mutant.

### t.Log does not fail a test

`t` has a third printing method, and it is easy to mix up with the others:

| Method | Prints a message | Marks the test failed |
|---|---|---|
| `t.Logf` | yes (shown with `-v` or when the test fails) | **no** |
| `t.Errorf` | yes | yes, and keeps going |
| `t.Fatalf` | yes | yes, and stops this test |

`t.Logf` is for extra detail while you investigate. A check that only calls
`t.Logf` passes no matter what the code does.

### Make sure your test can fail

Before you trust a new test, watch it fail once. Gopherlings' mutants do
that for you here. In your own projects, you can break the code on purpose
for a moment (return the wrong value, say), run `go test`, and check that the
test turns red.

## Your task

Deploy notifications show short commit IDs like `9f86d08`. `ShortSHA` in
`exercise.go` is correct:

```go
ShortSHA("9f86d081884c7d659a2feaa0c55ad015")
// => "9f86d08"   (the first 7 characters)

ShortSHA("a1b2c")
// => "a1b2c"     (7 characters or fewer come back unchanged)
```

A teammate wrote `TestShortSHA` in `exercise_test.go`, and it has two bugs.

### 1. Make the check able to fail

The check reports a mismatch with `t.Logf`, so it never fails. Change it to
`t.Errorf`.

After this step `TestShortSHA` fails, and `TestGrader` says your tests must
pass on the real code first. That is expected: the test now reports the
second bug.

### 2. Fix the expected value

Read the failure message. The test expects 8 characters, but `ShortSHA`
keeps 7. Change `want` to the correct short ID, `"9f86d08"`. Now the test
passes on the real code and catches the mutants that return the full ID or
keep 8 characters.

### 3. Check a short commit ID

Add a check that `ShortSHA("a1b2c")` returns `"a1b2c"`. Use `t.Errorf` if it
doesn't. This catches mutants that crash or return `""` when the ID is
shorter than 7 characters.

## Learn more

- [Package testing: T.Log, T.Error, T.Fatal](https://pkg.go.dev/testing#T)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
