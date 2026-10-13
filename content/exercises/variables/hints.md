## Task 1: join with +

Return one expression made of three parts joined with `+`: the `env`
parameter, the string `"/"`, and the `name` parameter. Order matters.

## Task 2: create ref with :=

`ref` doesn't exist yet, so create it with `:=`, then `return ref`. You need
five parts: `registry`, `"/"`, `name`, `":"`, `tag`.

## Task 3: copy the line above

The line `path = path + "/environments/" + env` shows the pattern. Write one
more line just like it, using `"/services/"` and `service`.

## Almost the answer

```go
func ServiceLabel(name, env string) string {
	return env + "/" + name
}

// inside ImageRef:
ref := registry + "/" + name + ":" + tag
return ref

// inside DeployPath:
path = path + "/services/" + service
```
