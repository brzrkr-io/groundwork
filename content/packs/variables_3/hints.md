## Task 1: Name a pod

A pod name is three parameters with a dash between each pair. Which parts are
parameters and which are fixed text?

### Closer

Fixed text goes in quotes (`"-"`); parameters don't. Join every piece with `+`:
`deployment + "-" + ...`.

### Answer

```go
return deployment + "-" + hash + "-" + suffix
```

## Task 2: Build an SSH target

This is one `return` line, like task 1: three parameters with separators
between them. Look at the example to see which separators.

### Closer

The separators are fixed text, so they go inside quotes: `"@"` and `"."`. The
parameters `user`, `host` and `domain` don't. Start with `user + "@" + ...`.

### Answer

```go
return user + "@" + host + "." + domain
```

## Task 3: Build a kubectl set-image command

This is a long string. It's easier to build in steps: create a variable, then
add to it with `+=`. Note that `deployment` appears twice.

### Closer

Fixed text, including its spaces, goes inside quotes; parameters don't:

```go
cmd := "kubectl -n " + namespace
cmd += " set image deployment/" + deployment
```

Then add a space, the deployment again, `"="`, and the image. Return `cmd`.

### Answer

```go
cmd := "kubectl -n " + namespace
cmd += " set image deployment/" + deployment
cmd += " " + deployment + "=" + image
return cmd
```
