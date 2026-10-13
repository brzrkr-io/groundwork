## Task 1: which slice?

`args` is the whole command line after `nodectl`, including the command name.
The command should only get what comes after its name. One slice expression
fixes it.

## Task 2: copy cordon

`runUncordon` is `runCordon` with three words changed. Then add a line to the
`commands` map literal, shaped exactly like the `"cordon"` line.

## Task 3: sorted names

Loop over the names in order and look each command up:

```go
for _, name := range slices.Sorted(maps.Keys(commands)) {
	fmt.Fprintf(w, "  %-9s %s\n", name, commands[name].summary)
}
```

Remember to add `"maps"` and `"slices"` to the import block.
