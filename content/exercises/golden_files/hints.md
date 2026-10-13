## Read, then compare

Task 1 has two steps: read the golden file into a variable with
`os.ReadFile(golden)`, then compare that with `got`. The `golden` variable
already holds the path. Remember to convert the bytes with `string(want)`.

## Task 1, nearly done

```go
want, err := os.ReadFile(golden)
if err != nil {
	t.Fatalf("reading %s: %v", golden, err)
}
if got != string(want) {
	t.Errorf("RenderInventory() output differs from %s\ngot:\n%s\nwant:\n%s", golden, got, want)
}
```

## Task 2: my test fails now

That's the point of task 2. Your test's "got:" output has a REGION column and
the "want:" output doesn't, because the golden file is outdated. Run
`go test -run TestRenderInventory -update` in a terminal inside the exercise
folder, check that `testdata/inventory.golden` now has three columns, and run
`go test` again.
