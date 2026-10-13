## Task 1: four cases

`Scalar` needs a case for `string`, one case listing `int, float64, bool`,
one for the interface `fmt.Stringer`, and a `default`. In the multi-type
case, `v` is still an `any`, which is exactly what `fmt.Sprintf("%v", v)`
needs.

## Task 2: replace the placeholder line

The starter stores every value with `%v`, even nested maps. Replace that line
with a type switch on `value`: a `case map[string]any:` that calls
`flattenInto(out, fullKey, v)`, and a `default:` that uses `Scalar`.

## Task 2: the error

When `Scalar` returns `false`, build the error with `fmt.Errorf`, using `%q`
for `fullKey` and `%T` for the value. Don't forget to return the error from
the recursive call too.

## Almost the answer

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
