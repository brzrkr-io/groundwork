# Table-driven tests: error cases

A parser is only as good as what it refuses. This time your table checks
both the inputs that should work and the ones that should fail, and you
write the whole test yourself.

## Learn

### Recap

As in *Table-driven tests*, a table test is a slice of anonymous structs,
one per case, plus a `for` loop that checks each row and reports failures
with `t.Errorf`, including the row's `name`. Test both sides of every
boundary. `TestGrader` again runs your test against mutants, and names any
that slip through.

### Rows that should fail: wantErr

When a function returns `(value, error)`, add a `wantErr bool` field. Rows
where an error is expected set it to `true` and leave `want` at `0`:

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

The loop then has two paths:

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

`continue` skips to the next row. In *Your first test* you used `t.Fatalf`
for an unexpected error. In a table that would stop the whole test, so the
other rows would never run. `t.Errorf` plus `continue` skips just this row.

Why not just compare `got` with `want` and skip `wantErr`? Because a broken
parser that returns `0, nil` for `"0Mi"` would match `want: 0` and pass.
Only checking the error catches it.

## Your task

The deploy tool reads container memory requests written the Kubernetes way
and converts them to MiB with `ParseMemory`. The code in `exercise.go` is
correct. The rules:

- The text is a whole number followed by a unit.
- The unit is `Mi` (MiB) or `Gi` (GiB). One `Gi` is `1024` MiB, not `1000`.
- Any other unit, or no unit at all, is an error.
- A missing or non-numeric number is an error.
- A number below `1` is an error.

```go
ParseMemory("512Mi")  // => 512, nil
ParseMemory("2Gi")    // => 2048, nil
ParseMemory("0Mi")    // => 0, error
```

`TestParseMemory` in `exercise_test.go` is empty. Fill it in. Pick any names
you like for the rows.

### 1. Write the table and the loop

Write a table with the fields `name`, `input`, `want` and `wantErr`, and a
loop that checks every row both ways, as in the Learn section. Start with
these rows:

| `input` | `want` | `wantErr` |
|---|---|---|
| `"512Mi"` | `512` | `false` |
| `"2Gi"` | `2048` | `false` |

This catches a mutant that multiplies `Gi` by `1000`.

### 2. Reject bad units

Add these rows, which must return an error:

| `input` | Why it fails |
|---|---|
| `"512"` | no unit |
| `"512Ki"` | `Ki` is not supported |

This catches mutants that accept a bare number or the `Ki` unit.

### 3. Reject bad numbers

Add these rows, which must also return an error:

| `input` | Why it fails |
|---|---|
| `"Gi"` | no number |
| `"lotsMi"` | not a number |
| `"0Mi"` | below 1 |
| `"-1Gi"` | below 1 |

This catches mutants that accept a missing or non-numeric number, zero, or a
negative number.

The exercise is done when `TestParseMemory` passes and `TestGrader` reports
that every mutant was caught.

## Learn more

- [Go wiki: Table-driven tests](https://go.dev/wiki/TableDrivenTests)
- [Kubernetes: Resource units for memory](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-memory)
- [Package testing](https://pkg.go.dev/testing)
