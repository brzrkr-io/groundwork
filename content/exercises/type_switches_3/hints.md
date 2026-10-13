## Task 1: Convert a plain value to text

`Scalar` needs a type switch with four branches: strings, the plain number/bool types, anything with a `String()` method, and everything else.

### Closer

Use `case string:`, then one case listing several types `case int, float64, bool:`, then `case fmt.Stringer:`, then `default:`. In the multi-type case `v` is still an `any`, which is exactly what `fmt.Sprintf("%v", v)` needs. In the `fmt.Stringer` case you can call `v.String()`.

### Answer

```go
switch v := v.(type) {
case string:
	return v, true
case int, float64, bool:
	return fmt.Sprintf("%v", v), true
case fmt.Stringer:
	return v.String(), true
default:
	return "", false
}
```

Cases are checked top to bottom, so plain types are handled before the `fmt.Stringer` check.

## Task 2: Flatten the config

The starter stores every value with `%v`, even nested maps. Replace that line so nested maps go through `flattenInto` again, and everything else goes through `Scalar`.

### Closer

Switch on `value`: a `case map[string]any:` that calls `flattenInto(out, fullKey, v)` and returns its error if it's not `nil`, and a `default:` that calls `Scalar(v)`. When `Scalar` returns `false`, build the error with `fmt.Errorf("unsupported value for %q: %T", fullKey, v)`.

### Answer

```go
switch v := value.(type) {
case map[string]any:
	if err := flattenInto(out, fullKey, v); err != nil {
		return err
	}
default:
	s, ok := Scalar(v)
	if !ok {
		return fmt.Errorf("unsupported value for %q: %T", fullKey, v)
	}
	out[fullKey] = s
}
```

Each nested map is handled by the same function one level deeper, with `fullKey` as its prefix.
