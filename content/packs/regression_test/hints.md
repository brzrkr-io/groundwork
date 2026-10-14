## Task 1: Reproduce bug report #1

Write a table test, as in *Table-driven tests: error cases*, with an input, the wanted name, and whether an error is wanted. Then add the cases from the bug report. The test is supposed to fail at this point.

### Closer

Use fields `name`, `input`, `want` and `wantErr`, and named fields in rows, like `{name: "capital P", input: "Production", want: "production"}`. Run each row with `t.Run(tc.name, ...)`, as in *Subtests and helpers*.

### Answer

```go
tests := []struct {
	name    string
	input   string
	want    string
	wantErr bool
}{
	{name: "lower case", input: "production", want: "production"},
	{name: "capital P", input: "Production", want: "production"},
	{name: "all caps", input: "STAGING", want: "staging"},
}
for _, tc := range tests {
	t.Run(tc.name, func(t *testing.T) {
		got, err := NormalizeEnvironment(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeEnvironment(%q) = %q, want an error", tc.input, got)
			}
			return
		}
		if err != nil {
			t.Fatalf("NormalizeEnvironment(%q) returned error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("NormalizeEnvironment(%q) = %q, want %q", tc.input, got, tc.want)
		}
	})
}
```

## Task 2: Fix bug report #1

Clean the name once at the top of `realNormalizeEnvironment`, then check the cleaned value and return it, not the original.

### Closer

`env := strings.ToLower(name)`, then `switch env` and `return env, nil`. Add `"strings"` to the imports.

### Answer

```go
env := strings.ToLower(name)
switch env {
case "development", "staging", "production":
	return env, nil
}
return "", fmt.Errorf("unknown environment %q", name)
```

## Task 3: Reproduce bug report #2

Add two more rows to the same table. Nothing in the loop changes.

### Closer

One input has a space at the end and the other has spaces on both sides plus a capital letter. Both expect the clean lower-case name.

### Answer

```go
{name: "trailing space", input: "staging ", want: "staging"},
{name: "spaces and caps", input: "  Development ", want: "development"},
```

## Task 4: Fix bug report #2

Spaces need removing before you lower-case and compare. You already have one cleaning line, so extend it.

### Closer

Wrap `name` in `strings.TrimSpace(...)` inside the `strings.ToLower` call.

### Answer

```go
env := strings.ToLower(strings.TrimSpace(name))
```

## Task 5: Don't over-fix

Add rows that must still fail. Use `wantErr: true` and leave `want` out.

### Closer

`{name: "...", input: "prod", wantErr: true}`. Do the same for the empty string `""`.

### Answer

```go
{name: "unknown name", input: "prod", wantErr: true},
{name: "empty", input: "", wantErr: true},
```
