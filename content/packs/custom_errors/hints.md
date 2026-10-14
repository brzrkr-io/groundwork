## Task 1: Write the error message

`Error` only has to build a string from the struct's fields. Inside the method, the receiver `e` gives you `e.Field` and `e.Reason`.

### Closer

You can join strings with `+`. Put the literal `": "` between the two fields.

### Answer

```go
return e.Field + ": " + e.Reason
```

## Task 2: Validate the replica count

Check the two bad cases with `if`, and return early from each one. If neither matches, the count is fine.

### Closer

Create the error with `&` so it is a `*ValidationError`, which is the type that has the `Error` method:

```go
if n < 1 {
	return &ValidationError{Field: "replicas", Reason: "must be at least 1"}
}
```

### Answer

```go
if n < 1 {
	return &ValidationError{Field: "replicas", Reason: "must be at least 1"}
}
if n > 50 {
	return &ValidationError{Field: "replicas", Reason: "must be at most 50"}
}
return nil
```

## Task 3: Pick an exit code

You need to look at what kind of error `err` holds: none, a `*ValidationError`, or something else. A type switch handles all three in one place.

### Closer

`switch err.(type) { ... }` lets each `case` name a type. `case nil` matches when there is no error, and `default` catches everything else.

### Answer

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
