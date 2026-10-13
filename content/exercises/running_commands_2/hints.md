## Task 1: one flag, two arguments

`append` takes any number of values. Give it `"-n"` and `namespace` as two
separate values. For the environment, start from `os.Environ()` (add `"os"` to
the imports):

```go
args = append(args, "-n", namespace)
cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
```

## Task 2: three fields, two buffers

Set `cmd.Stdin`, `cmd.Stdout` and `cmd.Stderr` before `cmd.Run()`. The output
fields need pointers: `&outBuf`, not `outBuf`.

## Task 2: almost the answer

```go
var outBuf, errBuf bytes.Buffer
cmd.Stdin = strings.NewReader(manifest)
cmd.Stdout = &outBuf
cmd.Stderr = &errBuf
err = cmd.Run()
return outBuf.String(), errBuf.String(), err
```

## Task 3: check "not found" first

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
