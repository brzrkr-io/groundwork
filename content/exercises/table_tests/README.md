# Table-driven tests

Real code has many cases worth checking. A *table-driven test* lists them as
rows of data and checks them all with one loop, so adding a case costs one
line. Most Go tests you'll read at work, including Kubernetes', look like
this.

**New here:** anonymous structs, slices of test cases, boundary testing.

## Learn

### The problem with copy and paste

In *Your first test*, every check repeated the same lines with different
values. With ten cases that's fifty lines, and a typo in one copy is easy to
miss. The only part that really changes is the data: the input and the
answer you want.

### A table of cases

Put the data in a slice of structs, one struct per case:

```go
tests := []struct {
	name string
	in   int
	want int
}{
	{"small", 2, 4},
	{"zero", 0, 0},
}
```

Read it in two halves:

1. `[]struct { name string; in int; want int }` is the **type**: a slice
   of structs with three fields.
2. The `{ ... }` after it holds the **rows**. Each row like
   `{"small", 2, 4}` fills the fields in order: `name`, `in`, `want`.

`name` describes the case in a few words. It doesn't change the check; it
tells you which row broke.

> **Why is it written like this?** Go requires the type before the `{`
> rows, and a comma after every row, even the last one when it ends a
> line. A `struct` with no type name is an *anonymous struct*: you need
> this type only here, so naming it is unnecessary. Calling the slice
> `tests` and the fields `name`/`want` is convention; the other field
> names (`in`) are your choice.

### One loop checks every row

```go
for _, tc := range tests {
	got := Double(tc.in)
	if got != tc.want {
		t.Errorf("%s: Double(%d) = %d, want %d", tc.name, tc.in, got, tc.want)
	}
}
```

`tc` ("test case") holds one row on each pass through the loop, and
`tc.in` reads a field from it. Use `t.Errorf`, not `t.Fatalf`, so one bad
row doesn't hide the rest. A failure reads like this:

```text
small: Double(2) = 5, want 4
```

Later, in *Subtests and helpers*, each row gets its own named subtest. For
now the name goes into the message.

> **Why is it written like this?** `range` gives an index and a value; the
> `_` throws the index away (Go rejects unused variables). The name `tc`
> is convention, and `tt` is also common.

### Test both sides of every boundary

Bugs gather at the edges. `>=` versus `>` changes the answer for exactly one
value, so a test that only checks values far from a limit can't tell them
apart. For every limit, test the last value before it and the first value
at it. For a rule like "adult from 18":

| Value | Why |
|---|---|
| 17 | last value before the limit |
| 18 | first value at the limit |

### How Gopherlings grades your test

As in *Your first test*, `TestGrader` runs your test against the real code
(it must pass) and against *mutants*, copies with one bug each (it must fail
on every one). Any mutant that slips through is named in `TestGrader`'s
output, such as "your tests didn't notice when DiskAlert(80) said "ok"".

## Your task

The monitoring agent turns disk usage into an alert level with `DiskAlert`.
The code in `exercise.go` is correct:

| `usedPercent` | Result |
|---|---|
| below 0 or above 100 | `"invalid"` |
| 90 to 100 | `"critical"` |
| 80 to 89 | `"warning"` |
| 0 to 79 | `"ok"` |

`TestDiskAlert` in `exercise_test.go` already has the table and the loop,
with one row, `{"half full", 50, "ok"}`. You only add rows to the table;
don't change the loop. Each row is name, `used`, `want`. Pick any names you
like.

### 1. Add a row for each level

**Goal:** add rows for `85` → `"warning"` and `95` → `"critical"`.

| `used` | `want` |
|---|---|
| `85` | `"warning"` |
| `95` | `"critical"` |

This catches mutants that never return `"warning"` or never return
`"critical"`.

**Think it through:**
- Look at the existing row. Which of its three values is the name, the input, the answer?
- Where do the new rows go: inside the type's `{ }` or the rows' `{ }`?

### 2. Add the boundary rows

**Goal:** add rows for both sides of the warning limit (80) and the critical limit (90).

| `used` | `want` |
|---|---|
| `79` | `"ok"` |
| `80` | `"warning"` |
| `89` | `"warning"` |
| `90` | `"critical"` |

This catches mutants that use `>` instead of `>=`, or start a level one
percent too early.

**Think it through:**
- For each value, find its line in the `DiskAlert` table above.
- Does each row end with a comma?

### 3. Add the edges of the valid range

**Goal:** add rows for the two ends of the valid range and the two values just outside it.

| `used` | `want` |
|---|---|
| `0` | `"ok"` |
| `100` | `"critical"` |
| `-1` | `"invalid"` |
| `101` | `"invalid"` |

This catches mutants that reject an empty or full disk, or accept `-1` or
`101`.

**Think it through:**
- `used` is an `int`, and `want` is a `string`. Which one needs quotes?

The exercise is done when `TestDiskAlert` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
- [Go by Example: Structs (anonymous structs)](https://gobyexample.com/structs)
