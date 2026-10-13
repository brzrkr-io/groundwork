## Task 1: join the strings

`Error` only has to build a string from the struct's fields. Inside the
method, the fields are `e.Field` and `e.Reason`. You can join strings with
`+`:

```go
return e.Field + ": " + e.Reason
```

## Task 2: return a pointer

Check the two bad cases with `if`, and return early from each one. Create the
error with `&` so it is a `*ValidationError`:

```go
if n < 1 {
	return &ValidationError{Field: "replicas", Reason: "must be at least 1"}
}
```

After both checks, finish with `return nil`.

## Task 3: switch on the type

A type switch has a case for each kind of error. `case nil` matches when
there is no error, and `default` catches everything else:

```go
switch err.(type) {
case nil:
	return 0
case *ValidationError:
	return 2
default:
	return 1
}
```
