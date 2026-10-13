# Running commands: stdin, stderr and PATH lookups

Real tools talk through all three streams: `kubectl apply -f -` reads a
manifest from stdin, prints results on stdout and warnings on stderr. This
time you wire those streams yourself, and check a tool is installed before
you rely on it.

## Learn

**Recap.** `exec.Command(name, args...)` starts a program directly, with no
shell. `Output` and `CombinedOutput` collect what it prints, a non-zero exit
gives an `*exec.ExitError`, and `cmd.Env` must start from `os.Environ()` or the
program loses `PATH`, `HOME` and the rest. The tests run the test binary
itself as a fake `kubectl` (the helper-process pattern from *Running
commands*), switched on by `TOOLRUN_HELPER=1`.

### Arguments are a slice, not a sentence

The program receives each slice element as one argument, exactly as written:

```go
exec.Command("kubectl", "get", "pods", "-n kube-system")   // 3 args: "-n kube-system" is ONE
exec.Command("kubectl", "get", "pods", "-n", "kube-system") // 4 args: correct
```

To add arguments conditionally, build a `[]string` and spread it with `...`:

```go
args := []string{"get", "pods"}
if namespace != "" {
	args = append(args, "-n", namespace)
}
cmd := exec.Command("kubectl", args...)
```

### Wiring the streams yourself

A `Cmd` has three fields for its streams:

| Field | Type | The program... |
|---|---|---|
| `cmd.Stdin` | `io.Reader` | reads its input from here |
| `cmd.Stdout` | `io.Writer` | writes normal output here |
| `cmd.Stderr` | `io.Writer` | writes errors and warnings here |

A `nil` field means "nothing": empty input, output thrown away. Set them,
then call `cmd.Run()`. A `bytes.Buffer` is an `io.Writer` that keeps
everything written to it; `buf.String()` gives it back.

```go
var stdout, stderr bytes.Buffer
cmd := exec.Command("kubectl", "apply", "-f", "-")
cmd.Stdin = strings.NewReader(manifest) // a string as an io.Reader
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()
```

Keeping the two buffers separate lets you show results and errors
differently, even when the command succeeds but prints a warning.

### Is the tool installed?

`exec.LookPath(name)` searches the directories in `PATH`, the same way your
shell does, and returns the full path of the program. If `name` contains a
slash, it checks that exact file instead. When nothing is found, the error
wraps `exec.ErrNotFound`:

```go
path, err := exec.LookPath("helm")
if errors.Is(err, exec.ErrNotFound) {
	fmt.Println("helm is not installed")
}
```

Other errors mean something else went wrong, for example the file exists but
isn't executable.

## Your task

Your platform team's deploy tool shells out to `kubectl`. In the examples,
`kubectl` is the fake kubectl from the tests (`os.Args[0]`).

### 1. Fix RolloutStatus

`RolloutStatus` has two bugs. The `-n` flag and the namespace are passed as
one argument, and `cmd.Env` drops our whole environment. Fix both: pass `"-n"`
and the namespace as separate arguments (only when `namespace` isn't `""`),
and set `cmd.Env` to `os.Environ()` plus `"KUBECONFIG=" + kubeconfig`.

```go
RolloutStatus(kubectl, "/tmp/config", "kube-system", "coredns")
// => `deployment "coredns" successfully rolled out`, nil
RolloutStatus(kubectl, "/tmp/config", "", "api")   // no -n flag at all
// => `deployment "api" successfully rolled out`, nil
```

### 2. Feed a manifest to kubectl apply

Finish `Apply`. Feed `manifest` to the command's stdin, collect stdout and
stderr in two separate buffers, and return both exactly as written (no
trimming) together with the error from `cmd.Run()`.

```go
Apply(kubectl, "kind: ConfigMap\nmetadata:\n  name: app-settings\n")
// => "configmap/app-settings configured\n",
//    "Warning: no namespace set, using \"default\"\n",
//    nil
Apply(kubectl, "")
// => "", "error: no objects passed to apply\n", <exit status 1 error>
```

### 3. Check that a tool is installed

Write `Installed` with `exec.LookPath`:

| `LookPath` result | `Installed` returns |
|---|---|
| found | `true, nil` |
| error wrapping `exec.ErrNotFound` | `false, nil` |
| any other error | `false, err` |

```go
Installed("fake-helm")                 // fake-helm is on PATH
// => true, nil
Installed("kubectl")                   // not on PATH
// => false, nil
Installed("/tmp/bin/notes.txt")        // exists, but not executable
// => false, <error>
```

## Learn more

- [exec.Cmd fields](https://pkg.go.dev/os/exec#Cmd)
- [exec.LookPath](https://pkg.go.dev/os/exec#LookPath)
- [kubectl apply](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_apply/)
