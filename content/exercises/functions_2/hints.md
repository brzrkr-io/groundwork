## Task 1: read the signature

`Percent(part, whole float64)` wants the smaller "part" first. Look at which
of `used` and `total` is the part.

## Task 2: a zero-value check

An empty string is the zero value of `string`. Check `if tag == ""`.

## Task 3: call the helper

```go
return registry + "/" + name + ":" + TagOrLatest(tag)
```
