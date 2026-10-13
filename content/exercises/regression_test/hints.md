## Start with the table

Your test needs a table with an input, the wanted name and whether an error
is wanted:

```go
tests := []struct {
	name    string
	input   string
	want    string
	wantErr bool
}{
	{name: "capital P", input: "Production", want: "production"},
}
```

Then loop over it with `t.Run(tc.name, ...)`, as in *Subtests and helpers*.

## The loop body

```go
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
```

## Fixing the code

Clean the name once, at the top of `realNormalizeEnvironment`, and switch on
the cleaned value:

```go
env := strings.ToLower(strings.TrimSpace(name))
```

Return `env`, not `name`, so the caller gets the lower-case form.
