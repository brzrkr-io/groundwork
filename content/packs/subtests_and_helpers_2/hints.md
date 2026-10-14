## Task 1: Write the helper and read a file

The helper has three steps: build a path inside `t.TempDir()`, write `content` there, and return the path. The subtest then calls `LoadHosts` on that path and checks the result.

### Closer

Build the path with `filepath.Join(t.TempDir(), "hosts.txt")`, write it with `os.WriteFile(path, []byte(content), 0o644)`, and put `t.Helper()` first. Wrap the check in `t.Run("reads hosts", func(t *testing.T) { ... })`.

### Answer

```go
func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hosts.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadHosts(t *testing.T) {
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
}
```

## Task 2: Check a file without a final newline

This subtest has the same shape as `"reads hosts"`. Only the name, the file content and `want` change.

### Closer

Copy the `"reads hosts"` subtest, rename it `"no trailing newline"`, and use the content `"web-1\ndb-1"`.

### Answer

```go
t.Run("no trailing newline", func(t *testing.T) {
	path := writeFile(t, "web-1\ndb-1")
	got, err := LoadHosts(path)
	if err != nil {
		t.Fatalf("LoadHosts returned error: %v", err)
	}
	want := []string{"web-1", "db-1"}
	if !slices.Equal(got, want) {
		t.Errorf("LoadHosts = %q, want %q", got, want)
	}
})
```

## Task 3: Check a missing file

Don't call `writeFile` here, because the file must not exist. Build the path yourself and check only the error.

### Closer

`filepath.Join(t.TempDir(), "missing.txt")` gives a path to a file that isn't there. Then `if _, err := LoadHosts(path); err == nil { ... }`.

### Answer

```go
t.Run("missing file", func(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.txt")
	if _, err := LoadHosts(path); err == nil {
		t.Errorf("LoadHosts(%q) returned no error, want an error", path)
	}
})
```
