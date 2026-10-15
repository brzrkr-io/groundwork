## Task 1: Fix the dispatch bug

`args` is everything after `nodectl`, including the command name. The command
should only get what comes after its name.

### Closer

One slice expression fixes it: `args[1:]` drops the first element.

### Answer

```go
return cmd.run(args[1:], stdout, stderr)
```

## Task 2: Add the uncordon command

`runUncordon` is `runCordon` with `cordon` changed to `uncordon`. Then add a
line to the `commands` map literal shaped like the `"cordon"` line.

### Closer

Rename the function, the FlagSet name and the messages to `uncordon`. The map
entry is `"uncordon": {summary: "...", run: runUncordon},`.

### Answer

```go
"uncordon": {summary: "allow new pods on a node again", run: runUncordon},

func runUncordon(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("uncordon", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "nodectl: uncordon needs exactly one NODE")
		return 2
	}
	fmt.Fprintf(stdout, "uncordoning %s\n", fs.Arg(0))
	return 0
}
```

## Task 3: Generate the help text

A plain `range` over a map visits the keys in random order. Loop over the
names sorted, and look each command up to get its summary.

### Closer

`slices.Sorted(maps.Keys(commands))` gives the names in order. Print each line
with `fmt.Fprintf(w, "  %-9s %s\n", ...)`. Add `"maps"` and `"slices"` to the
imports.

### Answer

```go
for _, name := range slices.Sorted(maps.Keys(commands)) {
	fmt.Fprintf(w, "  %-9s %s\n", name, commands[name].summary)
}
```
