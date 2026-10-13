## Start from Table-driven tests

Your table looks like the one in *Table-driven tests*, with one more field:
`wantErr bool`. Rows that should succeed set `wantErr` to `false`; rows that
should fail set it to `true` and put `0` in `want`.

## The loop has two paths

Inside the loop, call `ParseMemory(tc.input)`. If `tc.wantErr` is true, the
only thing to check is that `err` is not `nil`. Otherwise check that `err` is
`nil` and that `got` equals `tc.want`. Use `t.Errorf` and `continue` rather
than `t.Fatalf`, so one bad row doesn't stop the others.

## Nearly done

```go
tests := []struct {
	name    string
	input   string
	want    int
	wantErr bool
}{
	{"mebibytes", "512Mi", 512, false},
	{"no unit", "512", 0, true},
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

Add the remaining rows from the README.
