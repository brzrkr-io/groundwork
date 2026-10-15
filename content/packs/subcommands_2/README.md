# Subcommands: a command table and generated help

A `switch` is fine for three commands. Real tools have dozens, and each needs
a line in the help text. Here you drive a tool from a *command table*, so
adding a command means adding one entry.

**New here:** functions stored in struct fields, map lookup for dispatch. (Sorted map keys are from *Sorting*, `%-9s` padding from *Formatting with fmt*.)

## Learn

### Recap

As in *Subcommands: dispatching git-style commands*: `Run` receives
`os.Args[1:]`, so `args[0]` names the command and `args[1:]` belongs to it.
Each command parses its own arguments with its own `flag.FlagSet`. Asked-for
help goes to stdout with exit `0`; usage errors go to stderr with exit `2`.

### A table of commands

Instead of a `switch`, keep commands in a map from name to a struct. A struct
field can hold a function, if the function's signature matches the field's
type:

```go
type tool struct {
	about string
	do    func(n int) int
}

var tools = map[string]tool{
	"double": {about: "multiply by two", do: double},
}

t, ok := tools["double"]
if ok {
	fmt.Println(t.do(21)) // 42
}
```

> **Why is it written like this?** Go requires the signatures to match
> exactly, or it won't compile. `v, ok := m[key]` is Go's built-in way to ask
> "is it there?". Calling the second value `ok` is convention.

### Help text from the table

The table can also generate the help text, so help never forgets a command.
The catch: **map order is random** and changes between runs. As in
*Sorting*, loop over `slices.Sorted(maps.Keys(m))` to get keys in
alphabetical order.

### Padding with %-9s

To line text up in a column, give a verb a width. The number is a minimum
width; `-` puts the padding on the right:

```go
fmt.Printf("[%-6s]\n", "cat") // [cat   ]
fmt.Printf("[%6s]\n", "cat")  // [   cat]
```

### Duration flags

`runDrain` uses `fs.Duration` (as in *Flags: parsing options with a
FlagSet*). It accepts `90s` or `2m`, and prints with `%s` as `30s`, `2m0s`,
`1m30s`.

## Your task

`nodectl` helps with Kubernetes node maintenance. Before patching a node you
*cordon* it (no new pods), *drain* it (move its pods), then *uncordon* it.
Commands only print what they would do. Do the tasks in order: the help test
expects all three commands.

### 1. Fix the dispatch bug

**Goal:** make `Run` hand each command only its own arguments, not its name.

Someone reported:

```text
$ nodectl drain node-7
nodectl: drain needs exactly one NODE
```

The FlagSet sees `drain` as a positional argument. After the fix:

```text
$ nodectl drain node-7
draining node-7 (grace period 30s)
$ nodectl drain -grace 2m node-7
draining node-7 (grace period 2m0s)
$ nodectl cordon node-3
cordoning node-3
```

**Think it through:**
- For `nodectl drain node-7`, what's in `args`, index by index?
- Which part of `args` should the command see?

### 2. Add the uncordon command

**Goal:** write `runUncordon` and register it in `commands` as `"uncordon"`.

- Same signature as `runCordon`.
- Summary: `allow new pods on a node again`.
- No flags, but still give it a FlagSet named `"uncordon"`, so `-h` works and
  unknown flags are rejected.
- Exactly one NODE.

| Command line | stdout | stderr | Exit |
|---|---|---|---|
| `nodectl uncordon node-3` | `uncordoning node-3` | | `0` |
| `nodectl uncordon` | | `nodectl: uncordon needs exactly one NODE` | `2` |
| `nodectl uncordon node-3 node-4` | | `nodectl: uncordon needs exactly one NODE` | `2` |

**Think it through:**
- `runCordon` is almost the same. Which words change?
- In the map entry, `run:` gets the function's name, without `()`. Why?

### 3. Generate the help text

**Goal:** finish `writeUsage` so it lists every command, sorted by name.

After the three header lines it already writes, write one line per command:
two spaces, the name padded with `%-9s`, a space, then the summary, then a
newline. With all three commands, `nodectl help` prints exactly:

```text
usage: nodectl <command> [flags] NODE

commands:
  cordon    stop scheduling new pods on a node
  drain     evict all pods from a node
  uncordon  allow new pods on a node again
```

`Run` already uses `writeUsage` for `help`, `-h`, `--help` (stdout, exit
`0`), for no arguments (stderr, exit `2`), and after the
`nodectl: unknown command "reboot"` line (stderr, exit `2`). The test runs
each case 20 times, so an unsorted loop over the map gets caught. Add
`"maps"` and `"slices"` to the imports.

**Think it through:**
- Write to `w` (the parameter), not `stdout`: who decides where help goes?
- Loop over sorted names, then look up each summary by name.

## Learn more

- [Go by Example: Command-Line Subcommands](https://gobyexample.com/command-line-subcommands)
- [pkg.go.dev: maps.Keys](https://pkg.go.dev/maps#Keys) and [slices.Sorted](https://pkg.go.dev/slices#Sorted)
- [pkg.go.dev: fmt width and padding](https://pkg.go.dev/fmt#hdr-Printing)
- [Kubernetes: Safely drain a node](https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/)
