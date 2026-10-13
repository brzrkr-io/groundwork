## Task 1: two reader fields

The reader still uses the default comma and knows nothing about comments.
Look at the *Reader options* table: you need `Comma` and `Comment`. Both are
runes, so use single quotes: `'\t'` and `'#'`.

## Task 2: the usual writer ending

`Write` only fills a buffer. Replace `return nil` with a call to `cw.Flush()`
followed by `return cw.Error()`.

## Task 3: the loop

Copy the loop shape from `ReadRecords`, but ignore the record (`_, err :=
cr.Read()`). Before the `if err != nil` check, look for a field-count error
and `continue` so the loop keeps reading.

## Almost the answer (task 3)

```go
var pe *csv.ParseError
if errors.As(err, &pe) && errors.Is(pe.Err, csv.ErrFieldCount) {
	bad = append(bad, pe.Line)
	continue
}
if err != nil {
	return nil, err
}
```

Declare `var bad []int` before the loop and `return bad, nil` after it.
