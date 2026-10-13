# CSV files: fix the TSV DNS export

Not every "CSV" file uses commas. Many tools export *TSV* (tab-separated
values) with `#` comment lines. Here you fix a DNS zone export tool, then make
it point at the exact lines that are broken.

## Learn

### Recap

`csv.NewReader(r)` turns text into records (`[]string`), and `Read` returns
one record per call until `io.EOF`. `csv.NewWriter(w)` buffers what you
`Write`, so you must call `Flush()` and then check `Error()`. Without `Flush`,
small outputs come out empty.

### Reader options

A `*csv.Reader` has fields you set after `NewReader` and before the first
`Read`:

| Field | Default | Example | Effect |
|---|---|---|---|
| `Comma` | `','` | `cr.Comma = '\t'` | The field separator. A *rune*, so single quotes. |
| `Comment` | none | `cr.Comment = '#'` | Lines starting with this rune are skipped. |
| `FieldsPerRecord` | `0` | `cr.FieldsPerRecord = 3` | `0`: every record must match the first one. A positive number: every record must have exactly that many fields. `-1`: no check. |
| `TrimLeadingSpace` | `false` | `cr.TrimLeadingSpace = true` | Ignores spaces at the start of each field. |

`csv.Writer` has a `Comma` field too, so `cw.Comma = '\t'` writes TSV.
Blank lines are always skipped when reading.

### Errors that point at a line

When a record is bad, the reader returns a `*csv.ParseError`. Its `Error()`
text looks like `record on line 4: wrong number of fields`, and it has
fields you can read:

| Field | Meaning |
|---|---|
| `Line` | The line in the input where the problem is (comment and blank lines count too). |
| `Column` | The column (1-based) where the problem is. |
| `Err` | The underlying error, such as `csv.ErrFieldCount` or `csv.ErrBareQuote`. |

Use `errors.As` to get at those fields, as in *Custom error types*:

```go
var pe *csv.ParseError
if errors.As(err, &pe) && errors.Is(pe.Err, csv.ErrFieldCount) {
	fmt.Println("bad record on line", pe.Line)
}
```

A wrong field count is the one error you can keep going after: `Read` still
returns the record together with the error, and the next `Read` carries on
with the following line.

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

`ReadRecords` already reads every record into a `Record{Name, Type, Value}`,
but it fails on real exports with `record on line 1: wrong number of fields`.
The bug is in `newReader`: set it up for tab-separated fields and `#`
comments. Don't touch `FieldsPerRecord`.

```go
ReadRecords(strings.NewReader("# zone example.com\napi\tA\t10.0.0.1\n"))
// => []Record{{Name: "api", Type: "A", Value: "10.0.0.1"}}, nil
```

### 2. Fix the writer

`WriteRecords` should write each record as one tab-separated line, ending in
`\n`, with no header. Right now it writes nothing at all, and it always
returns `nil` even when the underlying writer fails. Make it flush the output
and return the writer's error.

```go
WriteRecords(w, []Record{
	{Name: "api", Type: "A", Value: "10.0.0.1"},
	{Name: "www", Type: "CNAME", Value: "api.example.com"},
})
// writes "api\tA\t10.0.0.1\nwww\tCNAME\tapi.example.com\n"
```

The tests also use a writer that always fails with the error `disk full`, and
check that `WriteRecords` returns that error (`errors.Is(err, errDiskFull)`).

### 3. Find the bad lines

Write `BadLines`. It reads an export using `newReader` (so it needs task 1),
and returns the line numbers of all records that don't have exactly three
fields, in order. It keeps reading after each bad record.

- A file with no bad records returns no line numbers and no error.
- Any other error, such as a stray `"` inside a field (`csv.ErrBareQuote`),
  stops the read: return `nil` and that error.

```go
in := "# zone example.com\napi\tA\t10.0.0.1\nwww\tCNAME\n\ndb\tA\t10.0.0.3\textra\n"
BadLines(strings.NewReader(in))
// => []int{3, 5}, nil   (line 1 is the comment, line 4 is blank)
```

Add `"errors"` to the import block for `errors.As`.

## Learn more

- [Package encoding/csv: Reader](https://pkg.go.dev/encoding/csv#Reader)
- [Package encoding/csv: ParseError](https://pkg.go.dev/encoding/csv#ParseError)
- [Package errors: As](https://pkg.go.dev/errors#As)
