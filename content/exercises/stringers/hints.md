## Task 1: one case per constant

Copy the `case Debug:` pattern for each of the other four constants. In the
`default` branch, build the text with `fmt.Sprintf` and `%d`. Convert the
receiver with `int(s)` first so `Sprintf` doesn't call `String` again.

## Task 2: the method signature

A `String` method on `Host` with a value receiver looks like this:

```go
func (h Host) String() string {
	// ...
}
```

Inside, `fmt.Sprintf` with `%s` for the name and IP and `%d` for the port
builds the text.

## Task 3: look at the receiver

The body of the `Disk` method is fine. The problem is `(d *Disk)`: only a
`*Disk` has that method, and the tests print a plain `Disk`. Remove the `*`.

## Almost the answer

```go
default:
	return fmt.Sprintf("Severity(%d)", int(s))
```

```go
func (h Host) String() string {
	return fmt.Sprintf("%s (%s:%d)", h.Name, h.IP, h.Port)
}
```

```go
func (d Disk) String() string {
```
