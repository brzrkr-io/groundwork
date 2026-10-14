## Task 1: Write the config into a temporary folder

Read the fixture file, turn it into a list of backends, and call `WriteUpstreamConfig` with a fresh temporary folder. Keep the folder in a variable, because you'll need it in Task 2.

### Closer

Build paths with `filepath.Join("testdata", "backends.txt")`. `strings.Fields(string(input))` splits the text into a `[]string`. Use `dir := t.TempDir()`, and stop with `t.Fatalf` on either error.

### Answer

```go
input, err := os.ReadFile(filepath.Join("testdata", "backends.txt"))
if err != nil {
	t.Fatalf("reading fixture: %v", err)
}
backends := strings.Fields(string(input))

dir := t.TempDir()
if err := WriteUpstreamConfig(dir, "api", backends); err != nil {
	t.Fatalf("WriteUpstreamConfig(%q, %q, %q) returned error: %v", dir, "api", backends, err)
}
```

## Task 2: Compare with the golden file

Read the file the function wrote, read the golden file, and compare them, as in *Test fixtures and golden files*.

### Closer

The written file is `filepath.Join(dir, "api.conf")`. Both reads give bytes, so compare with `string(got) != string(want)`.

### Answer

```go
got, err := os.ReadFile(filepath.Join(dir, "api.conf"))
if err != nil {
	t.Fatalf("reading written config: %v", err)
}

golden := filepath.Join("testdata", "api.conf.golden")
want, err := os.ReadFile(golden)
if err != nil {
	t.Fatalf("reading %s: %v", golden, err)
}
if string(got) != string(want) {
	t.Errorf("api.conf differs from %s\ngot:\n%s\nwant:\n%s", golden, got, want)
}
```

## Task 3: Add the -update flag

You need a package-level flag, and a step that rewrites the golden file from `got` when the flag is set. The order matters: write first, then read.

### Closer

Put `var update = flag.Bool("update", false, "rewrite golden files")` above the test. Inside, between reading `got` and reading `want`, add `if *update { ... os.WriteFile(golden, got, 0o644) ... }`.

### Answer

```go
var update = flag.Bool("update", false, "rewrite golden files")
```

```go
golden := filepath.Join("testdata", "api.conf.golden")
if *update {
	if err := os.WriteFile(golden, got, 0o644); err != nil {
		t.Fatalf("updating %s: %v", golden, err)
	}
}
want, err := os.ReadFile(golden)
```
