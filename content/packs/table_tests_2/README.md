# Table-driven tests: error cases

A parser is only as good as what it refuses. This time your table checks
both inputs that should work and inputs that should fail, and you write the
whole test yourself.

**New here:** a `wantErr` field, `continue` in a test loop.

## Learn

### Recap

As in *Table-driven tests*:

- A table is a slice of anonymous structs, one row per case.
- A `for` loop checks each row and reports failures with `t.Errorf`,
  including the row's `name`.
- Test both sides of every boundary.
- `TestGrader` runs your test against mutants and names any that slip
  through.

### Rows that should fail: wantErr

When a function returns `(value, error)`, add a `wantErr bool` field. Rows
that expect an error set it to `true` and put `0` in `want`:

```go
tests := []struct {
	name    string
	input   string
	want    int
	wantErr bool
}{
	{"ssh", "22", 22, false},
	{"not a number", "ssh", 0, true},
}
```

> **Why is it written like this?** `wantErr` is a community convention,
> not a Go rule; any `bool` field name works. The `0` in error rows is
> there because Go requires a value for every field when you fill a struct
> by position. Lining up the field types is what `gofmt` does for you.

### A loop with two paths

```go
for _, tc := range tests {
	got, err := ParsePort(tc.input)
	if tc.wantErr {
		if err == nil {
			t.Errorf("%s: ParsePort(%q) returned no error, want an error", tc.name, tc.input)
		}
		continue // got means nothing when there's an error
	}
	if err != nil {
		t.Errorf("%s: ParsePort(%q) returned error: %v", tc.name, tc.input, err)
		continue
	}
	if got != tc.want {
		t.Errorf("%s: ParsePort(%q) = %d, want %d", tc.name, tc.input, got, tc.want)
	}
}
```

Read it top to bottom:

1. Rows that **should** fail: complain if there was no error, then move on.
2. Rows that **should** work: complain about an unexpected error, then move on.
3. Only now compare `got` with `want`.

`continue` jumps straight to the next row.

> **Why is it written like this?** In *Your first test* you used
> `t.Fatalf` for an unexpected error. In a table that stops the whole
> test, so later rows never run. `t.Errorf` plus `continue` skips just
> this row. That's a convention you'll see everywhere in Go code.

Why not skip `wantErr` and just compare `got` with `want`? A broken parser
that returns `0, nil` for bad input would match `want: 0` and pass. Only
checking the error catches it.

## Your task

The deploy tool reads container memory requests written the Kubernetes way
and converts them to MiB with `ParseMemory`. The code in `exercise.go` is
correct. The rules:

- The text is a whole number followed by a unit.
- The unit is `Mi` (MiB) or `Gi` (GiB). One `Gi` is `1024` MiB, not `1000`.
- Any other unit, or no unit, is an error.
- A missing or non-numeric number is an error.
- A number below `1` is an error.

```go
ParseMemory("512Mi")  // => 512, nil
ParseMemory("2Gi")    // => 2048, nil
ParseMemory("0Mi")    // => 0, error
```

`TestParseMemory` in `exercise_test.go` is empty. You write all of it. Pick
any names you like for the rows.

### 1. Write the table and the loop

**Goal:** in `TestParseMemory`, write a table with fields `name`, `input`, `want`, `wantErr`, plus a loop that checks every row both ways.

Start with these rows:

| `input` | `want` | `wantErr` |
|---|---|---|
| `"512Mi"` | `512` | `false` |
| `"2Gi"` | `2048` | `false` |

This catches a mutant that multiplies `Gi` by `1000`.

**Think it through:**
- Which function do you call in the loop, and which field do you pass it?
- What type is each field? `input` is text; `want` is a number.
- Which two `continue`s does the loop need, and why?

### 2. Reject bad units

**Goal:** add rows for `"512"` and `"512Ki"`, which must both return an error.

| `input` | Why it fails |
|---|---|
| `"512"` | no unit |
| `"512Ki"` | `Ki` is not supported |

This catches mutants that accept a bare number or the `Ki` unit.

**Think it through:**
- What goes in `want` when there's no valid answer?
- What goes in `wantErr`?

### 3. Reject bad numbers

**Goal:** add rows for `"Gi"`, `"lotsMi"`, `"0Mi"` and `"-1Gi"`, which must all return an error.

| `input` | Why it fails |
|---|---|
| `"Gi"` | no number |
| `"lotsMi"` | not a number |
| `"0Mi"` | below 1 |
| `"-1Gi"` | below 1 |

This catches mutants that accept a missing or non-numeric number, zero, or a
negative number.

**Think it through:**
- These rows look just like the ones in task 2. Which value changes per row?

The exercise is done when `TestParseMemory` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [Kubernetes: Resource units for memory](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory)
- [Package testing](https://pkg.go.dev/testing)
