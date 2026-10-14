## Task 1: Fix the reader settings

The reader still splits on commas and knows nothing about comments. Look at
the *Reader options* table in the README.

### Closer

Set two fields on `cr`: `Comma` and `Comment`. Both are runes, so use single
quotes: `'\t'` and `'#'`.

### Answer

```go
cr.Comma = '\t'
cr.Comment = '#'
```

## Task 2: Fix the writer

`Write` only fills a buffer, so nothing reaches `w` until you flush. Errors
from writing are kept inside the writer until you ask for them.

### Closer

Replace `return nil` with a call to `cw.Flush()` followed by returning
`cw.Error()`.

### Answer

```go
cw.Flush()
return cw.Error()
```

## Task 3: Find the bad lines

Copy the loop shape from `ReadRecords`, but you don't need the record itself.
A wrong field count comes back as a `*csv.ParseError`, which knows its line
number, and the reader can carry on after it.

### Closer

Use `_, err := cr.Read()`, stop on `io.EOF`. Before the usual `err != nil`
check, use `errors.As(err, &pe)` with `var pe *csv.ParseError`, and check
`errors.Is(pe.Err, csv.ErrFieldCount)`. If so, append `pe.Line` and
`continue`. Add `"errors"` to the imports.

### Answer

```go
cr := newReader(r)
var bad []int
for {
	_, err := cr.Read()
	if err == io.EOF {
		break
	}
	var pe *csv.ParseError
	if errors.As(err, &pe) && errors.Is(pe.Err, csv.ErrFieldCount) {
		bad = append(bad, pe.Line)
		continue
	}
	if err != nil {
		return nil, err
	}
}
return bad, nil
```

Only field-count errors are skipped; anything else, like a bare quote, still
stops the read.
