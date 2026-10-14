## Task 1: Build the service label

You need three pieces glued together: the `env` parameter, a slash, then the
`name` parameter. Which piece comes first?

### Closer

Fixed text goes inside quotes (`"/"`); parameters don't (`env`, `name`). Join
the pieces with `+`: `return env + ...`.

### Answer

```go
return env + "/" + name
```

## Task 2: Build the image reference

`ref` doesn't exist yet, so you create it with `:=`, then return it. It is
made of five pieces: `registry`, a slash, `name`, a colon, then `tag`.

### Closer

Fixed text goes inside quotes (`"/"` and `":"`); parameters don't. Create the
variable, then return it:

```go
ref := registry + "/" + ...
return ref
```

### Answer

```go
ref := registry + "/" + name + ":" + tag
return ref
```

## Task 3: Finish the deploy API path

Look at the line `path = path + "/environments/" + env`. You need one more line
just like it, for the service.

### Closer

`path` already exists, so change it with `=` (not `:=`). Fixed text goes inside
quotes (`"/services/"`); the parameter `service` doesn't:
`path = path + "/services/" + ...`.

### Answer

```go
path = path + "/services/" + service
```
