# Subtests and helpers

A test with many cases gets hard to read when every case is one long block of
`if` statements. *Subtests* give each case its own name and result, and
*helpers* move the repeated checking code out of the way. Together they make a
test you can read like a list of rules.

## Learn

### Subtests with t.Run

`t.Run(name, func(t *testing.T) { ... })` runs a small test inside your test.
The function you pass gets its own `t`, so a failure is reported under the
subtest's name.

```go
func TestParseHosts(t *testing.T) {
	t.Run("skips comments", func(t *testing.T) {
		// checks for this case
	})
	t.Run("trims spaces", func(t *testing.T) {
		// checks for this case
	})
}
```

`go test -v` lists each subtest separately:

```text
--- FAIL: TestParseHosts (0.00s)
    --- PASS: TestParseHosts/skips_comments (0.00s)
    --- FAIL: TestParseHosts/trims_spaces (0.00s)
```

Go replaces spaces in subtest names with `_`. You can run a single subtest
by putting its name after a `/`:

```text
go test -run 'TestParseHosts/skips_comments'
```

A `t.Fatalf` inside a subtest stops only that subtest. The others still run.

### Subtests and tables

Subtests fit table-driven tests, from *Table-driven tests*, perfectly. Give
each row a `name` and call `t.Run` in the loop:

```go
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		got := ParseHosts(tt.text)
		// compare got with tt.want
	})
}
```

Adding a case is now just adding a row, and the output tells you which row
failed by name.

### Comparing slices

You can't compare two slices with `==`. The `slices` package has
`slices.Equal(a, b)`, which returns `true` when both have the same length and
the same elements in the same order.

```go
slices.Equal([]string{"web-1"}, []string{"web-1"}) // true
slices.Equal([]string{"web-1"}, []string{"db-1"})  // false
slices.Equal(nil, []string{})                      // true
```

The last line matters: a `nil` slice and an empty slice both have length 0,
so `slices.Equal` treats them as equal. Your test doesn't care which one
`ParseHosts` returns for "no hosts".

### Test helpers

When several tests do the same check, move it into a *helper* function that
takes `t`:

```go
func assertPort(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
```

Calling `t.Helper()` first tells Go "this is a helper". When the check
fails, Go then reports the line in the test that *called* the helper, not the
line inside the helper. Without it, every failure would point at the same
`t.Errorf` line, and you'd have to guess which call failed. `t.Helper()`
changes only the message, not whether the test passes.

### How Gopherlings grades your test

As in *Your first test*, `TestGrader` re-runs `TestParseHosts` against the
real `ParseHosts`, where it must pass, and then against *mutants*, broken
copies of `ParseHosts`. Your test must fail against every mutant. If one
slips through, `TestGrader` names the bug your test missed.

## Your task

The deploy tool reads its target list from a hosts file: one hostname per
line, with comments and blank lines for humans. `ParseHosts` in
`exercise.go` is already correct. It trims the spaces around each line, skips
blank lines and lines starting with `#`, and returns the hosts in order.

```go
ParseHosts("# web\nweb-1\n\n  web-2  \n")
// => []string{"web-1", "web-2"}
```

`TestParseHosts` in `exercise_test.go` already has a table with one case,
`"one host per line"`, and a loop that runs each row as a subtest.

### 1. Finish the helper

`assertHosts` is empty, so right now it never fails, and the example case
catches nothing. Make it:

- call `t.Helper()`,
- compare `got` and `want` with `slices.Equal` (add `"slices"` to the
  imports),
- fail with `t.Errorf("got %q, want %q", got, want)` if they differ.

Now the example case catches a `ParseHosts` that drops the last host.
Gopherlings can't check that you called `t.Helper()`, since it only improves
the failure messages, but do it anyway.

### 2. Add the cases

Add three rows to the table, named exactly like this:

| `name` | `text` | `want` |
|---|---|---|
| `"skips comments"` | `"# web tier\nweb-1\n# database\ndb-1"` | `[]string{"web-1", "db-1"}` |
| `"skips blank lines"` | `"web-1\n\nweb-2\n\n"` | `[]string{"web-1", "web-2"}` |
| `"trims spaces"` | `"  web-1  \n\tdb-1"` | `[]string{"web-1", "db-1"}` |

`\n` is a newline and `\t` is a tab. Each row catches a mutant that forgets
one of the rules.

The exercise is done when `TestParseHosts` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Using subtests and sub-benchmarks (Go blog)](https://go.dev/blog/subtests)
- [testing.T.Run](https://pkg.go.dev/testing#T.Run)
- [testing.T.Helper](https://pkg.go.dev/testing#T.Helper)
- [slices.Equal](https://pkg.go.dev/slices#Equal)
