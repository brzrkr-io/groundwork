## Task 1: Remove the shadowing field

`Key` lives on `ObjectMeta`, so it only sees `ObjectMeta`'s `Name`. The extra
`Name` on `Secret` takes the value instead. Keep only one `Name`.

### Closer

Delete the `Name string` line from `Secret`. The compiler then points at
`NewSecret`: move `name` into the `ObjectMeta{...}` literal.

### Answer

```go
type Secret struct {
	ObjectMeta
	Data map[string]string
}

func NewSecret(namespace, name string, data map[string]string) Secret {
	return Secret{
		ObjectMeta: ObjectMeta{Name: name, Namespace: namespace},
		Data:       data,
	}
}
```

## Task 2: Make Validate check the metadata too

`Secret.Validate` replaces the promoted one, so `ObjectMeta`'s checks never
run. Call the inner method yourself at the top, before the data check.

### Closer

Reach the inner method through the embedded field: `s.ObjectMeta.Validate()`.
Don't write `s.Validate()`: that is the method you are in, so it would call
itself forever. Store the result in `err` and return it if it isn't `nil`.

### Answer

```go
func (s Secret) Validate() error {
	err := s.ObjectMeta.Validate()
	if err != nil {
		return err
	}
	if len(s.Data) == 0 {
		return errors.New("secret has no data")
	}
	return nil
}
```
