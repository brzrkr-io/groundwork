# CSV files: fix the TSV DNS export

Not every "CSV" file uses commas. Many tools export *TSV* (tab-separated
values) with `#` comment lines. Here you fix a DNS zone export tool, then make
it point at the exact lines that are broken.

**New here:** `csv.Reader` options, `Read` in a loop, `*csv.ParseError`.

## Learn

### Recap

As in *CSV files: a host inventory*: `csv.NewReader(r)` turns text into
records (`[]string`). `csv.NewWriter(w)` buffers what you `Write`; nothing
reaches `w` until `Flush()`, and `Error()` reports what went wrong.

### Reading one record at a time

`Read` returns one record per call. When the input runs out it returns
`io.EOF`, the normal end, not a failure:

```go
for {
	record, err := cr.Read()
	if err == io.EOF {
		break
	}
	if err != nil {
		return err
	}
	fmt.Println(record[0])
}
```

### Reader options

A `*csv.Reader` has fields you set after `NewReader` and before the first
`Read`:

| Field | Default | Example | Effect |
|---|---|---|---|
| `Comma` | `','` | `cr.Comma = ';'` | The field separator. |
| `Comment` | none | `cr.Comment = '!'` | Lines starting with this character are skipped. |
| `FieldsPerRecord` | `0` | `cr.FieldsPerRecord = 4` | `0`: every record must match the first. Positive: exactly that many fields. `-1`: no check. |
| `TrimLeadingSpace` | `false` | `cr.TrimLeadingSpace = true` | Ignores spaces at the start of each field. |

`csv.Writer` has a `Comma` field too. Blank lines are always skipped when
reading.

> **Why is it written like this?** `Comma` and `Comment` are *runes* (single
> characters), so Go requires single quotes: `';'`, not `";"`. A tab is
> written `'\t'`. Setting options as plain fields, rather than passing them
> to `NewReader`, is the package's design.

### Errors that point at a line

When a record is bad, the reader returns a `*csv.ParseError`. Its text looks
like `record on line 4: wrong number of fields`, and it has fields you can
read:

| Field | Meaning |
|---|---|
| `Line` | The input line with the problem (comment and blank lines count too). |
| `Column` | The column (1-based) with the problem. |
| `Err` | The underlying error, such as `csv.ErrFieldCount` or `csv.ErrBareQuote`. |

Use `errors.As` to reach those fields, as in *Custom error types*:

```go
var pe *csv.ParseError
if errors.As(err, &pe) && errors.Is(pe.Err, csv.ErrBareQuote) {
	fmt.Println("stray quote at column", pe.Column)
}
```

> **Why is it written like this?** `errors.As` needs a pointer to your
> variable so it can fill it in, so Go requires `&pe`. The `&&` means the
> second check only runs if the first one found a `ParseError`.

A wrong field count is the one error you can keep going after: the next
`Read` carries on with the following line. Other errors, like a stray quote,
mean you should stop.

## Your task

Your DNS provider exports each zone as a tab-separated file. Each record line
has three fields: name, type and value. There's no header row, and lines
starting with `#` are comments. (In the examples, `\t` is a tab character.)

```text
# zone example.com
api	A	10.0.0.1
www	CNAME	api.example.com
```

### 1. Fix the reader settings

**Goal:** make `newReader` return a reader that splits fields on tabs and
skips `#` comment lines.

`ReadRecords` already uses it, but fails on real exports with
`record on line 1: wrong number of fields`. Leave `FieldsPerRecord` as it is.

```go
ReadRecords(strings.NewReader("# zone example.com\napi\tA\t10.0.0.1\n"))
// => []Record{{Name: "api", Type: "A", Value: "10.0.0.1"}}, nil
```

**Think it through:**
- Which two rows of the options table match "tabs" and "comments"?
- These are runes. Which kind of quotes do they need?

### 2. Fix the writer

**Goal:** make `WriteRecords` actually send its output to `w` and return the
writer's error.

Each record is one tab-separated line ending in `\n`, with no header. Right
now nothing reaches `w`, and it returns `nil` even when `w` fails.

```go
WriteRecords(w, []Record{
	{Name: "api", Type: "A", Value: "10.0.0.1"},
	{Name: "www", Type: "CNAME", Value: "api.example.com"},
})
// writes "api\tA\t10.0.0.1\nwww\tCNAME\tapi.example.com\n"
```

The tests also use a writer that always fails with the error `disk full`,
and check that `WriteRecords` returns that error.

**Think it through:**
- Check the Recap: what pushes buffered output to `w`?
- Which method reports errors from earlier `Write` calls? Return what it gives you instead of `nil`.

### 3. Find the bad lines

**Goal:** `BadLines(r)` returns the line numbers of every record that doesn't
have exactly three fields, in order.

- Read with `newReader` (so task 1 must be done), one record at a time,
  until `io.EOF`. Keep reading after each bad record.
- No bad records → `nil` and no error.
- Any other error, such as a stray `"` inside a field (`csv.ErrBareQuote`),
  stops the read: return `nil` and that error.
- Add `"errors"` to the import block.

```go
in := "# zone example.com\napi\tA\t10.0.0.1\nwww\tCNAME\n\ndb\tA\t10.0.0.3\textra\n"
BadLines(strings.NewReader(in))
// => []int{3, 5}, nil   (line 1 is the comment, line 4 is blank)
```

**Think it through:**
- Which `ParseError` field holds the line number, and which tells you it's a field-count problem?
- Check for the field-count case *before* the general `err != nil`. What do you do after recording the line?

## Learn more

- [Package encoding/csv: Reader](https://pkg.go.dev/encoding/csv#Reader)
- [Package encoding/csv: ParseError](https://pkg.go.dev/encoding/csv#ParseError)
- [Package errors: As](https://pkg.go.dev/errors#As)
