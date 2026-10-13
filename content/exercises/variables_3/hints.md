## Task 1 and 2: one return line

Both are a single `return` that joins parameters and separators with `+`.
For `PodName` the separators are `"-"`; for `SSHTarget` they are `"@"` and
`"."`.

## Task 3: build it in steps

Long lines are easier with a variable you grow step by step:

```go
cmd := "kubectl -n " + namespace
cmd += " set image deployment/" + deployment
```

Then add a space, the deployment name again, `"="`, and the image.

## Almost the answer

```go
func SetImageCommand(namespace, deployment, image string) string {
	cmd := "kubectl -n " + namespace
	cmd += " set image deployment/" + deployment
	cmd += " " + deployment + "=" + image
	return cmd
}
```
