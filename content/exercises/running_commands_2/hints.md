## Task 1: Fix RolloutStatus

Each element of the args slice reaches the program as one argument, so
`"-n " + namespace` arrives as a single odd word. Also, setting `cmd.Env`
replaces the whole environment unless you start from ours.

### Closer

`append` takes any number of values, so pass `"-n"` and `namespace`
separately. Start the environment from `os.Environ()` (add `"os"` to the
imports).

### Answer

```go
args := []string{"rollout", "status", "deployment/" + deployment}
if namespace != "" {
	args = append(args, "-n", namespace)
}
cmd := exec.Command(kubectl, args...)
cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
```

## Task 2: Feed a manifest to kubectl apply

`exec.Cmd` has three fields for the streams: `Stdin`, `Stdout` and `Stderr`.
Set them before `cmd.Run()`, then read the buffers afterwards.

### Closer

`strings.NewReader(manifest)` is a reader for stdin. A `bytes.Buffer` is a
writer, but the fields need a pointer: `&outBuf`, not `outBuf`.

### Answer

```go
var outBuf, errBuf bytes.Buffer
cmd.Stdin = strings.NewReader(manifest)
cmd.Stdout = &outBuf
cmd.Stderr = &errBuf
err = cmd.Run()
return outBuf.String(), errBuf.String(), err
```

## Task 3: Check that a tool is installed

`exec.LookPath` only gives you an error to inspect. Check for the "not found"
case first, before treating any error as a real problem.

### Closer

`errors.Is(err, exec.ErrNotFound)` is true when the program simply isn't
there. You only need the error from `LookPath`, so discard the path with `_`.

### Answer

```go
_, err := exec.LookPath(name)
if errors.Is(err, exec.ErrNotFound) {
	return false, nil
}
if err != nil {
	return false, err
}
return true, nil
```
