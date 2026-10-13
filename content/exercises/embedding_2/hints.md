## Task 1: one Name, in one place

`Key` lives on `ObjectMeta`, so it only sees `ObjectMeta`'s `Name`. Delete the
`Name string` line from `Secret`. The compiler then points at `NewSecret`: move
`name` into the `ObjectMeta{...}` literal.

## Task 2: reach the inner method

Inside `Secret.Validate`, `s.ObjectMeta.Validate()` runs the metadata checks.
Don't write `s.Validate()`: that is the method you are in, so it would call
itself forever.

## Task 2: return the error early

Store the result in `err`. If it isn't `nil`, return it straight away. Only
then check `Data`.

## Almost the answer

```go
// Task 1
return Secret{
	ObjectMeta: ObjectMeta{Name: name, Namespace: namespace},
	Data:       data,
}

// Task 2, at the top of Secret.Validate
err := s.ObjectMeta.Validate()
if err != nil {
	return err
}
```
