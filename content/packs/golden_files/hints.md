## Task 1: Compare with the golden file

There are two steps: read the golden file into a variable, then compare it with `got`. The `golden` variable already holds the path.

### Closer

`want, err := os.ReadFile(golden)` gives you bytes, so compare with `got != string(want)`. Stop with `t.Fatalf` if `err != nil`.

### Answer

```go
want, err := os.ReadFile(golden)
if err != nil {
	t.Fatalf("reading %s: %v", golden, err)
}
if got != string(want) {
	t.Errorf("RenderInventory() output differs from %s\ngot:\n%s\nwant:\n%s", golden, got, want)
}
```

## Task 2: Update the outdated golden file

Your test is supposed to fail after Task 1. Look at the failure: the "got:" output has a REGION column and the "want:" output doesn't, because the golden file is outdated.

### Closer

You don't change any code. The test already has an `-update` branch that rewrites the golden file from `got`. Run it from a terminal in the exercise folder.

### Answer

```text
go test -run TestRenderInventory -update
```

Then check that `testdata/inventory.golden` has three columns (NAME, IP, REGION) and run `go test` again.
