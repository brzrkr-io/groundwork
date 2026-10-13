## Task 1: the helper

Three steps: `filepath.Join(t.TempDir(), "hosts.txt")` for the path,
`os.WriteFile(path, []byte(content), 0o644)` to write it, and `return path`.
Put `t.Helper()` first.

## Task 1, nearly done

```go
func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hosts.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
```

```go
t.Run("reads hosts", func(t *testing.T) {
	path := writeFile(t, "# web tier\nweb-1\n\n  web-2\n")
	got, err := LoadHosts(path)
	if err != nil {
		t.Fatalf("LoadHosts returned error: %v", err)
	}
	want := []string{"web-1", "web-2"}
	if !slices.Equal(got, want) {
		t.Errorf("LoadHosts = %q, want %q", got, want)
	}
})
```

## Task 2: same shape, new content

Copy the `"reads hosts"` subtest, rename it, and change the content and
`want`. Nothing else changes.

## Task 3: no file at all

Don't call `writeFile` here. Build the path yourself and only check the
error:

```go
path := filepath.Join(t.TempDir(), "missing.txt")
if _, err := LoadHosts(path); err == nil {
	t.Errorf("LoadHosts(%q) returned no error, want an error", path)
}
```
