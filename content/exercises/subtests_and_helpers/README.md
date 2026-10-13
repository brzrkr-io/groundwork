# Subtests and helpers

A test with many cases gets hard to read when every case is one long block
of `if` statements. *Subtests* give each case its own name and result.
*Helpers* move repeated checking code out of the way. Together they make a
test you can read like a list of rules.

**New here:** `t.Run`, `t.Helper`, `slices.Equal`, struct rows with field names.

## Learn

### Subtests with t.Run

`t.Run(name, func(t *testing.T) { ... })` runs a small test inside your
test. The function you pass gets its own `t`, so a failure is reported under
the subtest's name.

```go
func TestParseSize(t *testing.T) {
	t.Run("gigabytes", func(t *testing.T) {
		// checks for this case
	})
	t.Run("no unit", func(t *testing.T) {
		// checks for this case
	})
}
```

`go test -v` lists each subtest separately. Go replaces spaces in subtest
names with `_`:

```text
--- FAIL: TestParseSize (0.00s)
    --- PASS: TestParseSize/gigabytes (0.00s)
    --- FAIL: TestParseSize/no_unit (0.00s)
```

To run one subtest, put its name after a `/`:
`go test -run 'TestParseSize/no_unit'`.

A `t.Fatalf` inside a subtest stops only that subtest. The others still run.

> **Why is it written like this?** The `func(t *testing.T) { ... }` is a
> function with no name, written right where it's used. Go requires that
> exact parameter type. Naming the inner parameter `t` again is
> convention: inside the subtest you always use its own `t`.

### Subtests and tables

Subtests fit table-driven tests (from *Table-driven tests*) perfectly: call
`t.Run` in the loop, with the row's `name`:

```go
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		got := ParseSize(tt.text)
		// compare got with tt.want
	})
}
```

### Rows with field names

Rows can name their fields instead of relying on order:

```go
{
	name: "gigabytes",
	text: "2GB",
	want: 2048,
},
```

> **Why is it written like this?** Go allows both styles. With names, the
> order doesn't matter and long values stay readable on their own lines.
> Every line ends with a comma; Go requires that when the closing `}` is on
> the next line. One field per line is the author's choice.

### Comparing slices

You can't compare two slices with `==`. The `slices` package has
`slices.Equal(a, b)`: `true` when both have the same length and the same
elements in the same order.

```go
slices.Equal([]string{"a"}, []string{"a"}) // true
slices.Equal([]string{"a"}, []string{"b"}) // false
slices.Equal(nil, []string{})              // true
```

The last line matters: a `nil` slice and an empty slice both have length 0,
so `slices.Equal` treats them as equal. To use it, add `"slices"` to the
`import` block.

### Test helpers

When several checks do the same thing, move that into a *helper* function
that takes `t`:

```go
func assertPort(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
```

`t.Helper()` tells Go "this is a helper". When the check fails, Go then
reports the line that *called* the helper, not the `t.Errorf` line inside
it. Without it, every failure points at the same line and you have to guess
which call failed. It changes only the message, not whether the test passes.

> **Why is it written like this?** Taking `t *testing.T` as the first
> parameter is what lets a helper fail the test; that's convention. Putting
> `t.Helper()` on the first line is also convention. Names like
> `assertPort` are your choice.

### How Gopherlings grades your test

As in *Your first test*, `TestGrader` runs your test against the real code
(it must pass) and against *mutants*, copies with one bug each (it must fail
on every one). It names any mutant that slips through.

## Your task

The deploy tool reads its target list from a hosts file: one hostname per
line, with comments and blank lines for humans. `ParseHosts` in
`exercise.go` is correct. It trims spaces around each line, skips blank
lines and lines starting with `#`, and returns the hosts in order.

```go
ParseHosts("# web\nweb-1\n\n  web-2  \n")
// => []string{"web-1", "web-2"}
```

`TestParseHosts` in `exercise_test.go` already has a table with one case,
`"one host per line"`, and a loop that runs each row as a subtest and calls
the helper `assertHosts`.

### 1. Finish the helper

**Goal:** make `assertHosts` fail the test when `got` and `want` differ.

`assertHosts` is empty, so right now it never fails. Make it:

- call `t.Helper()`,
- compare `got` and `want` with `slices.Equal` (add `"slices"` to the
  imports),
- if they differ, fail with `t.Errorf`, showing both slices with `%q` in a
  "got …, want …" message.

Then the existing case catches a `ParseHosts` that drops the last host.
Gopherlings can't check that you called `t.Helper()`, since it only improves
the messages, but do it anyway.

**Think it through:**
- `slices.Equal` returns `true` when they match. When should the test fail?
- The file has `import "testing"`. What does it look like with two imports?

### 2. Add the cases

**Goal:** add three rows to the table in `TestParseHosts`, named exactly as below.

| `name` | `text` | `want` |
|---|---|---|
| `"skips comments"` | `"# web tier\nweb-1\n# database\ndb-1"` | `[]string{"web-1", "db-1"}` |
| `"skips blank lines"` | `"web-1\n\nweb-2\n\n"` | `[]string{"web-1", "web-2"}` |
| `"trims spaces"` | `"  web-1  \n\tdb-1"` | `[]string{"web-1", "db-1"}` |

`\n` is a newline and `\t` is a tab, written inside the quotes. Each row
catches a mutant that forgets one rule. Don't change the loop; it already
runs every row.

**Think it through:**
- Look at the `"one host per line"` row. Which three field names does it use?
- Where does one row end and the next begin (look for `},`)?

The exercise is done when `TestParseHosts` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Using subtests and sub-benchmarks (Go blog)](https://go.dev/blog/subtests)
- [testing.T.Run](https://pkg.go.dev/testing#T.Run)
- [testing.T.Helper](https://pkg.go.dev/testing#T.Helper)
- [slices.Equal](https://pkg.go.dev/slices#Equal)
