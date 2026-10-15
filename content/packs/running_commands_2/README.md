# Running commands: stdin, stderr and PATH lookups

`kubectl apply -f -` reads a manifest from stdin, prints results on stdout
and warnings on stderr. This time you wire those streams yourself, and check a
tool is installed before relying on it.

**New here:** `cmd.Stdin`/`Stdout`/`Stderr`, `bytes.Buffer`, `exec.LookPath`.

## Learn

### Recap

`exec.Command(name, args...)` starts a program directly, no shell. A non-zero
exit gives an `*exec.ExitError`. `cmd.Env` must start from `os.Environ()`, or
the program loses `PATH`, `HOME` and the rest. The tests run the test binary
as a fake `kubectl`, switched on by `TOOLRUN_HELPER=1`.

### Arguments are a slice, not a sentence

Each string you pass reaches the program as **one** argument:

```go
exec.Command("ls", "-l -a")    // 1 argument: ls sees "-l -a" and fails
exec.Command("ls", "-l", "-a") // 2 arguments: correct
```

To add arguments only sometimes, build a `[]string` and spread it with `...`:

```go
args := []string{"log"}
if limit != "" {
	args = append(args, "--max-count", limit)
}
cmd := exec.Command("git", args...)
```

> **Why is it written like this?** The `...` after `args` is required: it
> tells Go "pass these as separate arguments". `append` takes as many values
> as you like after the slice.

### Wiring the streams yourself

| Field | Type | The program… |
|---|---|---|
| `cmd.Stdin` | `io.Reader` | reads its input from here |
| `cmd.Stdout` | `io.Writer` | writes normal output here |
| `cmd.Stderr` | `io.Writer` | writes errors and warnings here |

A `nil` field means empty input or thrown-away output. Set the fields, then
call `cmd.Run()`. A `bytes.Buffer` keeps everything written to it;
`buf.String()` gives it back. `strings.NewReader(s)` turns a string into a
reader.

```go
var result bytes.Buffer
cmd := exec.Command("sort")
cmd.Stdin = strings.NewReader("b\na\n")
cmd.Stdout = &result
err := cmd.Run()
// result.String() == "a\nb\n"
```

> **Why is it written like this?** `&result` (with `&`) is required: `Write`
> is defined on `*bytes.Buffer`, so only a pointer is an `io.Writer`. `var x
> bytes.Buffer` is fine because its zero value is an empty buffer, ready to use.

### Is the tool installed?

`exec.LookPath(name)` searches `PATH` like your shell does, and returns the
full path. If `name` contains a slash, it checks that exact file. When
nothing is found, the error wraps `exec.ErrNotFound`:

```go
_, err := exec.LookPath("terraform")
if errors.Is(err, exec.ErrNotFound) {
	fmt.Println("please install terraform")
}
```

Other errors mean something else went wrong, like a file that exists but
isn't executable.

## Your task

Your platform team's deploy tool shells out to `kubectl`. In the examples,
`kubectl` is the fake one from the tests (`os.Args[0]`).

### 1. Fix RolloutStatus

**Goal:** fix the two bugs in `RolloutStatus` so the fake kubectl accepts
the call.

- Bug 1: `-n` and the namespace are joined into one argument. Pass `"-n"`
  and the namespace as **two** arguments, only when `namespace` isn't `""`.
- Bug 2: `cmd.Env` drops our whole environment. Make it `os.Environ()` plus
  `"KUBECONFIG="` followed by the `kubeconfig` parameter. Add `"os"` to the
  imports.

```go
RolloutStatus(kubectl, "/tmp/config", "kube-system", "coredns")
// => `deployment "coredns" successfully rolled out`, nil
RolloutStatus(kubectl, "/tmp/config", "", "api")   // no -n flag at all
// => `deployment "api" successfully rolled out`, nil
```

**Think it through:**
- Both bugs have a `TODO(1)` right above them. Change those two lines.
- In bug 2, which part is fixed text in quotes and which is the parameter?

### 2. Feed a manifest to kubectl apply

**Goal:** make `Apply` send `manifest` to stdin and return stdout and stderr
separately, plus the error from `cmd.Run()`.

Collect each stream in its own buffer. Return both strings exactly as
written: no trimming. Add `"bytes"` to the imports.

```go
Apply(kubectl, "kind: ConfigMap\nmetadata:\n  name: app-settings\n")
// => "configmap/app-settings configured\n",
//    "Warning: no namespace set, using \"default\"\n",
//    nil
Apply(kubectl, "")
// => "", "error: no objects passed to apply\n", <exit status 1 error>
```

**Think it through:**
- Set the three fields **before** `cmd.Run()`; read the buffers **after**.
- `err` is already declared in the function line, so use `=`, not `:=`.

### 3. Check that a tool is installed

**Goal:** write `Installed(name)` with `exec.LookPath`.

| `LookPath` result | `Installed` returns |
|---|---|
| found | `true, nil` |
| error wrapping `exec.ErrNotFound` | `false, nil` |
| any other error | `false, err` |

Add `"errors"` to the imports.

```go
Installed("fake-helm")            // fake-helm is on PATH
// => true, nil
Installed("kubectl")              // not on PATH
// => false, nil
Installed("/tmp/bin/notes.txt")   // exists, but not executable
// => false, <error>
```

**Think it through:**
- You only need `LookPath`'s error. What do you write for a value you ignore?
- Check the "not found" case before the general `err != nil` case.

## Learn more

- [exec.Cmd fields](https://pkg.go.dev/os/exec#Cmd)
- [exec.LookPath](https://pkg.go.dev/os/exec#LookPath)
- [kubectl apply](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_apply/)
