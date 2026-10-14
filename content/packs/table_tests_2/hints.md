## Task 1: Write the table and the loop

Your table looks like the one in *Table-driven tests*, with one more field, `wantErr bool`. The loop has two paths: rows that should fail, and rows that should succeed.

### Closer

In the loop, call `got, err := ParseMemory(tc.input)`. If `tc.wantErr` is true, only check that `err != nil`, then `continue`. Otherwise check that `err == nil` and that `got == tc.want`. Use `t.Errorf` and `continue`, not `t.Fatalf`, so one bad row doesn't stop the rest.

### Answer

```go
tests := []struct {
	name    string
	input   string
	want    int
	wantErr bool
}{
	{"mebibytes", "512Mi", 512, false},
	{"gibibytes", "2Gi", 2048, false},
}

for _, tc := range tests {
	got, err := ParseMemory(tc.input)
	if tc.wantErr {
		if err == nil {
			t.Errorf("%s: ParseMemory(%q) returned no error, want an error", tc.name, tc.input)
		}
		continue
	}
	if err != nil {
		t.Errorf("%s: ParseMemory(%q) returned error: %v", tc.name, tc.input, err)
		continue
	}
	if got != tc.want {
		t.Errorf("%s: ParseMemory(%q) = %d, want %d", tc.name, tc.input, got, tc.want)
	}
}
```

## Task 2: Reject bad units

These rows must fail, so set `wantErr` to `true`. What goes in `want` when there's no valid answer?

### Closer

Error rows put `0` in `want` and `true` in `wantErr`, like `{"some name", "512", 0, true}`.

### Answer

```go
{"no unit", "512", 0, true},
{"unsupported unit", "512Ki", 0, true},
```

## Task 3: Reject bad numbers

Same pattern as Task 2: four more rows that must return an error.

### Closer

Add a row for each of `"Gi"`, `"lotsMi"`, `"0Mi"` and `"-1Gi"`, each with `0, true`.

### Answer

```go
{"no number", "Gi", 0, true},
{"not a number", "lotsMi", 0, true},
{"zero", "0Mi", 0, true},
{"negative", "-1Gi", 0, true},
```
