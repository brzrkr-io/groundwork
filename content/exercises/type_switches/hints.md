## Task 1: Pick out alerts

You need to ask `e` whether it holds an `AlertEvent`, without crashing when it doesn't. The comma-ok form of a type assertion does exactly that.

### Closer

`e.(AlertEvent)` gets the `AlertEvent` out of `e`. Put two variables on the left to get an `ok` result instead of a panic: `alert, ok := e.(AlertEvent)`. Then return both.

### Answer

```go
alert, ok := e.(AlertEvent)
return alert, ok
```

When `e` holds something else (or `nil`), `alert` is the zero `AlertEvent{}` and `ok` is `false`.

## Task 2: Describe any event

Use a type switch with one `case` per event type, and keep the starter's `return` as the `default`. Inside each case you can read that type's fields.

### Closer

Bind the value with `switch v := e.(type)`. Inside `case ScaleEvent:`, `v` is a `ScaleEvent`, so `v.From` and `v.To` work. Build each line with `fmt.Sprintf`, using `%s` for strings and `%d` for ints:

```go
switch v := e.(type) {
case DeployEvent:
	return fmt.Sprintf("deploy %s %s", v.Service, v.Version)
// case ScaleEvent: ...
// case AlertEvent: ...
default:
	return fmt.Sprintf("unknown event %T", e)
}
```

### Answer

```go
switch v := e.(type) {
case DeployEvent:
	return fmt.Sprintf("deploy %s %s", v.Service, v.Version)
case ScaleEvent:
	return fmt.Sprintf("scale %s %d -> %d", v.Service, v.From, v.To)
case AlertEvent:
	return fmt.Sprintf("alert %s [%s]: %s", v.Service, v.Severity, v.Message)
default:
	return fmt.Sprintf("unknown event %T", e)
}
```

A pointer like `*DeployEvent` matches none of the cases, so it lands in `default`.

## Task 3: Tally a batch of events

The switch is already there. Add a case for each event type that adds one to the right key, just like `default` does for `"unknown"`.

### Closer

You only need the type, not the value, so `switch e.(type)` is enough. Each case is one line, for example `case ScaleEvent: counts["scale"]++`.

### Answer

```go
switch e.(type) {
case DeployEvent:
	counts["deploy"]++
case ScaleEvent:
	counts["scale"]++
case AlertEvent:
	counts["alert"]++
default:
	counts["unknown"]++
}
```

`counts[key]++` works on a missing key because it starts from the zero value `0`.
