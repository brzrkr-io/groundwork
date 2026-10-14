## Task 1: Name the severities

The `case Debug:` branch already shows the pattern. Each other constant needs
its own `case`, and the `default` branch needs to put the number into the text.

### Closer

Copy `case Debug: return "DEBUG"` once per constant. In `default`, use
`fmt.Sprintf` with `%d`, and pass `int(s)` rather than `s`, so `Sprintf`
doesn't call `String` again and loop forever.

### Answer

```go
switch s {
case Debug:
	return "DEBUG"
case Info:
	return "INFO"
case Warning:
	return "WARNING"
case Error:
	return "ERROR"
case Critical:
	return "CRITICAL"
default:
	return fmt.Sprintf("Severity(%d)", int(s))
}
```

## Task 2: Print a host

`fmt` looks for a method called exactly `String` that takes nothing and
returns a `string`. Write one on `Host` with a value receiver, then build the
text from the three fields.

### Closer

The method header is `func (h Host) String() string`. Inside, `fmt.Sprintf`
with `%s` for the name and IP and `%d` for the port builds the text:
`"%s (%s:%d)"`.

### Answer

```go
func (h Host) String() string {
	return fmt.Sprintf("%s (%s:%d)", h.Name, h.IP, h.Port)
}
```

## Task 3: Fix the disk printout

The body of the `Disk` method is already right. Look at the receiver: which
type actually has this method, `Disk` or `*Disk`?

### Closer

With `(d *Disk)`, only a `*Disk` has `String`, and the tests print a plain
`Disk`. Change the receiver to a value receiver.

### Answer

```go
func (d Disk) String() string {
	return fmt.Sprintf("%s %d%% used", d.Mount, d.UsedPercent)
}
```

A value receiver gives the method to both `Disk` and `*Disk`, so `fmt` finds it either way.
