# Your first test: fix a broken test

Tests have bugs too. A test that can never fail is worse than no test at
all, because it makes you feel safe when you aren't.

**New here:** `t.Logf`, checking that a test can fail.

## Learn

### Recap

As in *Your first test*:

- A test is a `func TestXxx(t *testing.T)` in a `_test.go` file.
- It calls your code, compares *got* with *want*, and reports a mismatch.
- `TestGrader` runs your test against the real code (it must pass) and
  against *mutants*, copies with one bug each (it must fail on every one).

### t.Logf does not fail a test

`t` has a third printing method, and it is easy to mix up with the others:

| Method | Prints a message | Marks the test failed |
|---|---|---|
| `t.Logf` | yes (shown with `-v`, or when the test fails anyway) | **no** |
| `t.Errorf` | yes | yes, and keeps going |
| `t.Fatalf` | yes | yes, and stops this test |

`t.Logf` is for extra detail while you investigate. A check that only calls
`t.Logf` passes no matter what the code does.

```go
if got != want {
	t.Logf("note: got %d", got) // prints, but the test still passes
}
```

> **Why is it written like this?** All three methods take the same
> arguments (a format string, then values), so swapping one for another
> changes only the method name. Which one you pick is your choice, and it
> decides what the test does on a mismatch.

### Make sure your test can fail

Before you trust a new test, watch it fail once. Here Groundwork's mutants
do that for you. In your own projects, break the code on purpose for a moment
(return a wrong value, say), run `go test`, and check that the test turns
red.

### Read the failure message

A good failure message says what was called, what came back, and what was
expected:

```text
Double(4) = 9, want 8
```

When a test fails, first ask: is the code wrong, or is the `want` wrong?
Here the code is correct, so a wrong `want` is a bug in the test.

## Your task

Deploy notifications show short commit IDs. `ShortSHA` in `exercise.go` is
correct. It keeps the first 7 characters; IDs of 7 characters or fewer come
back unchanged.

```go
ShortSHA("9f86d081884c7d659a2feaa0c55ad015")
// => "9f86d08"   (the first 7 characters)

ShortSHA("a1b2c")
// => "a1b2c"     (5 characters, so unchanged)
```

A teammate wrote `TestShortSHA` in `exercise_test.go`, and it has two bugs.
Fix them, then add one more check.

### 1. Make the check able to fail

**Goal:** make the existing check in `TestShortSHA` mark the test failed when `got != want`.

Right now the `if got != want` block calls a method that only prints. Keep
the message and arguments; change only which method is called.

After this step `TestShortSHA` fails, and `TestGrader` says your tests must
pass on the real code first. That is expected: the test now shows you the
second bug.

**Think it through:**
- In the table above, which methods mark the test failed?
- Should the test stop at this mismatch, or keep going to later checks?

### 2. Fix the expected value

**Goal:** change `want` in `TestShortSHA` so it matches what the correct `ShortSHA` returns for the long ID.

Read the failure message from step 1. The correct value is in the first
example above. Once `want` is right, the test passes on the real code and
catches mutants that return the full ID or keep 8 characters.

**Think it through:**
- How many characters does `ShortSHA` keep? How many does `want` have?
- Is it the code or the test that is wrong here?

### 3. Check a short commit ID

**Goal:** add a check to `TestShortSHA` that `ShortSHA("a1b2c")` returns `"a1b2c"`.

Report a mismatch with `t.Errorf`. This catches mutants that crash or return
`""` when the ID is shorter than 7 characters.

**Think it through:**
- The check above already has the got/want shape. Which parts change?
- In the message, which value is the input and which is the expected result?

## Learn more

- [Package testing: T.Log, T.Error, T.Fatal](https://pkg.go.dev/testing#T)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
