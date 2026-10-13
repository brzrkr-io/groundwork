## Task 1: the comma-ok assertion

`e.(AlertEvent)` gets the `AlertEvent` out of `e`. Add a second variable on
the left to get the `ok` result instead of a panic:

```go
alert, ok := e.(AlertEvent)
```

Then return both values.

## Task 2: one case per event type

Bind the value with `switch v := e.(type)`. Inside `case ScaleEvent:`, `v` is
a `ScaleEvent`, so `v.From` and `v.To` work. Build each line with
`fmt.Sprintf`, using `%s` for strings and `%d` for ints.

## Task 3: count by kind

The switch is already there. Add a case for each event type that adds one to
the right key, just like the `default` case does for `"unknown"`.

## Almost the answer

```go
switch v := e.(type) {
case DeployEvent:
	return fmt.Sprintf("deploy %s %s", v.Service, v.Version)
case ScaleEvent:
	return fmt.Sprintf("scale %s %d -> %d", v.Service, v.From, v.To)
// case AlertEvent: ... "alert %s [%s]: %s"
default:
	return fmt.Sprintf("unknown event %T", e)
}
```

For `Tally`, each case is one line, for example `case ScaleEvent: counts["scale"]++`.
