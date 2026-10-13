# Subcommands: a command table and generated help

A `switch` is fine for three commands. Real tools have dozens, and each one
needs a line in the help text. Here you drive a tool from a *command table*,
so adding a command means adding one entry.

## Learn

**Recap.** `Run` receives `os.Args[1:]`, so `args[0]` names the command and
`args[1:]` belongs to it. Each command parses its own arguments with its own
`flag.FlagSet`. Help that was asked for goes to stdout with exit code `0`.
Usage errors go to stderr with exit code `2`.

### A table of commands

Instead of a `switch`, keep the commands in a map from name to a struct. A
struct field can hold a function, as long as the function's signature matches
the field's type:

```go
type command struct {
	summary string
	run     func(args []string, stdout, stderr io.Writer) int
}

var commands = map[string]command{
	"cordon": {summary: "stop scheduling new pods on a node", run: runCordon},
}
```

Dispatch is then a map lookup, and calling the stored function looks like any
other call:

```go
cmd, ok := commands[args[0]]
if !ok {
	// unknown command
}
return cmd.run(args[1:], stdout, stderr)
```

### Help text from the table

The same table can generate the help text, so the help never forgets a
command. There is a catch: **map order is random**, and it changes from run to
run. Help text that shuffles itself is confusing and breaks tests. As in
*Sorting*, loop over `slices.Sorted(maps.Keys(commands))` to get the names in
alphabetical order.

To line the summaries up in a column, pad the name with `%-9s`. The `9` is a
minimum width and the `-` puts the padding on the right:

```go
fmt.Printf("[%-9s]\n", "drain") // [drain    ]
```

### Duration flags

`runDrain` uses `fs.Duration` (seen in *Flags: parsing options with a FlagSet*). It accepts values like `90s` or
`2m`, and a `time.Duration` prints itself with `%s` as `30s`, `2m0s` or
`1m30s`.

## Your task

`nodectl` is a small helper for Kubernetes node maintenance. Before you patch
a node, you *cordon* it (no new pods), *drain* it (move its pods elsewhere),
and afterwards *uncordon* it. The commands only print what they would do. Do
the tasks in order: the help test expects all three commands.

### 1. Fix the dispatch bug

Someone reported this:

```text
$ nodectl drain node-7
nodectl: drain needs exactly one NODE
```

`Run` hands each command the wrong slice, so the FlagSet sees `drain` as a
positional argument. Fix it so that:

```text
$ nodectl drain node-7
draining node-7 (grace period 30s)
$ nodectl drain -grace 2m node-7
draining node-7 (grace period 2m0s)
$ nodectl cordon node-3
cordoning node-3
```

### 2. Add the uncordon command

Write `runUncordon` with the same signature as `runCordon`, and register it in
`commands` under the name `"uncordon"` with the summary
`allow new pods on a node again`. It takes no flags (still give it a FlagSet
named `"uncordon"`, so `-h` works and unknown flags are rejected) and exactly
one NODE:

| Command line | stdout | stderr | Exit |
|---|---|---|---|
| `nodectl uncordon node-3` | `uncordoning node-3` | | `0` |
| `nodectl uncordon` | | `nodectl: uncordon needs exactly one NODE` | `2` |
| `nodectl uncordon node-3 node-4` | | `nodectl: uncordon needs exactly one NODE` | `2` |

### 3. Generate the help text

Finish `writeUsage`. After the three header lines it already writes, write one
line per command, sorted by name: two spaces, the name padded with `%-9s`, a
space, and the summary. With all three commands registered, `nodectl help`
prints exactly:

```text
usage: nodectl <command> [flags] NODE

commands:
  cordon    stop scheduling new pods on a node
  drain     evict all pods from a node
  uncordon  allow new pods on a node again
```

`Run` already uses `writeUsage` for `help`, `-h`, `--help` (stdout, exit `0`),
for no arguments (stderr, exit `2`), and after the
`nodectl: unknown command "reboot"` line (stderr, exit `2`). The test calls
each one 20 times, so a loop straight over the map will be caught. Add
`"maps"` and `"slices"` to the imports.

## Learn more

- [Go by Example: Command-Line Subcommands](https://gobyexample.com/command-line-subcommands)
- [maps.Keys](https://pkg.go.dev/maps#Keys) and [slices.Sorted](https://pkg.go.dev/slices#Sorted)
- [fmt: width and padding](https://pkg.go.dev/fmt#hdr-Printing)
- [Kubernetes: Safely drain a node](https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/)
