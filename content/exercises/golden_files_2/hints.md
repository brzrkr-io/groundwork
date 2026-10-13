## Same steps as before, plus a file

Read the fixture, call the function, read the file it wrote, then compare it
with the golden file exactly as in *Test fixtures and golden files*. Every
file path is built with `filepath.Join`, for example
`filepath.Join(dir, "api.conf")`.

## Task 1, nearly done

```go
input, err := os.ReadFile(filepath.Join("testdata", "backends.txt"))
if err != nil {
	t.Fatalf("reading fixture: %v", err)
}
backends := strings.Fields(string(input))

dir := t.TempDir()
if err := WriteUpstreamConfig(dir, "api", backends); err != nil {
	t.Fatalf("WriteUpstreamConfig returned error: %v", err)
}
```

## Task 2, nearly done

```go
got, err := os.ReadFile(filepath.Join(dir, "api.conf"))
if err != nil {
	t.Fatalf("reading written config: %v", err)
}
golden := filepath.Join("testdata", "api.conf.golden")
want, err := os.ReadFile(golden)
// ... check err, then compare string(got) with string(want)
```

## Task 3, nearly done

Above the test: `var update = flag.Bool("update", false, "rewrite golden files")`.
Inside it, between reading `got` and reading `want`:

```go
if *update {
	if err := os.WriteFile(golden, got, 0o644); err != nil {
		t.Fatalf("updating %s: %v", golden, err)
	}
}
```
