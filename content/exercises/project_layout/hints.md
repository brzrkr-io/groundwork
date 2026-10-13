## Tasks 1 and 2: copy from the old main

The old `main` already has both pieces of logic. In `Status`, check the
biggest threshold first (`>= 90`, then `>= 80`). In `Parse`, the
`strings.Cut` and `strconv.Atoi` lines move over almost unchanged. Replace
each `fmt.Fprintf(os.Stderr, …)` and `os.Exit(2)` with `return Reading{},
fmt.Errorf(…)`, without the `diskcheck: ` prefix and the newline.

## Task 3: Fprintf to the writer

```go
var readings []disk.Reading
for _, arg := range args {
	r, err := disk.Parse(arg)
	if err != nil {
		// task 4
	}
	readings = append(readings, r)
}
code := 0
for _, r := range readings {
	status := disk.Status(r.Percent)
	if status == "full" {
		code = 1
	}
	fmt.Fprintf(stdout, "%s %d%% %s\n", r.Host, r.Percent, status)
}
return code
```

## Task 4: return early

The no-arguments check at the top of `run` shows the shape. Inside the parse
loop, do the same with the error:

```go
fmt.Fprintf(stderr, "diskcheck: %v\n", err)
return 2
```
