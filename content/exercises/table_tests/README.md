# Table-driven tests

Real code has many cases worth checking. A *table-driven test* lists them as
rows of data and checks them all with one loop, so adding a case costs one
line. Most Go tests you'll read at work, including Kubernetes', look like
this.

## Learn

### The problem with copy and paste

In *Your first test*, every check repeated the same five lines with a
different port. With ten cases that's fifty lines, and a typo in one copy is
easy to miss. The only part that really changes is the data: the input and
the answer you want.

### A table of cases

Put the data in a slice of structs, one struct per case:

```go
tests := []struct {
	name string
	port string
	want int
}{
	{"ssh", "22", 22},
	{"https", "443", 443},
}
```

`struct { ... }` without a type name is an *anonymous struct*. You only need
this type here, so it doesn't get a name. Each `{"ssh", "22", 22}` fills in
the fields in order: `name`, `port`, `want`.

The `name` field describes the case in a few words. It doesn't affect the
check. It's there so a failure tells you which row broke.

### One loop checks every row

```go
for _, tc := range tests {
	got := Double(tc.in)
	if got != tc.want {
		t.Errorf("%s: Double(%d) = %d, want %d", tc.name, tc.in, got, tc.want)
	}
}
```

`tc` is short for "test case". It holds one row on each pass through the
loop. Use `t.Errorf`, not `t.Fatalf`, so one bad row doesn't hide the rest.
A failure then reads like this:

```text
warning starts: DiskAlert(80) = "ok", want "warning"
```

Later, in *Subtests and helpers*, you'll give each row its own named subtest
with `t.Run`. For now the name goes into the message.

### Test both sides of every boundary

Bugs gather at the edges. `>=` versus `>` changes the answer for exactly one
value, so a test that only checks 50 and 95 can't tell them apart. For every
limit in the rules, add the first value on each side:

| Rule | Rows to add |
|---|---|
| warning from 80 | 79 and 80 |
| critical from 90 | 89 and 90 |
| valid from 0 to 100 | -1, 0, 100 and 101 |

### How Gopherlings grades your test

As in *Your first test*, `TestGrader` in `grader_test.go` re-runs your
`TestDiskAlert` against the real `DiskAlert`, where it must pass, and then
against *mutants*, copies with one deliberate bug each. Your test must fail
against every mutant. Any mutant that slips through is named in
`TestGrader`'s output, for example "your tests didn't notice when
DiskAlert(80) said "ok"".

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
with one row, `{"half full", 50, "ok"}`. You only add rows. Pick any names
you like.

### 1. Add a row for each level

Add these rows:

| `used` | `want` |
|---|---|
| `85` | `"warning"` |
| `95` | `"critical"` |

This catches mutants that never return `"warning"` or never return
`"critical"`.

### 2. Add the boundary rows

Add both sides of the warning and critical boundaries:

| `used` | `want` |
|---|---|
| `79` | `"ok"` |
| `80` | `"warning"` |
| `89` | `"warning"` |
| `90` | `"critical"` |

This catches mutants that use `>` instead of `>=`, or start a level one
percent too early.

### 3. Add the edges of the valid range

| `used` | `want` |
|---|---|
| `0` | `"ok"` |
| `100` | `"critical"` |
| `-1` | `"invalid"` |
| `101` | `"invalid"` |

This catches mutants that reject an empty or full disk, or accept `-1` or
`101`.

The exercise is done when `TestDiskAlert` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [Go by Example: Testing and Benchmarking](https://gobyexample.com/testing-and-benchmarking)
- [Go by Example: Structs (anonymous structs)](https://gobyexample.com/structs)
