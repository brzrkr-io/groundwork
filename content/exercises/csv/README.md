# CSV files: a host inventory

Spreadsheets, cloud billing exports, asset inventories and monitoring tools
all speak *CSV* (comma-separated values). Go's `encoding/csv` package reads
and writes it correctly, so you can turn an inventory export into Go structs
and back again.

## Learn

### Why not `strings.Split`?

A CSV line looks easy to split on commas, until a field contains a comma:

```text
hostname,region,owner
web-1,eu-west-1,"platform, sre"
```

`strings.Split` would give you four pieces for the second line. CSV wraps a
field in double quotes when it contains a comma, a quote or a newline, and
writes a quote inside a field as two quotes (`"say ""hi"""`). The `csv`
package handles all of that for you.

### Reading

`csv.NewReader` takes any `io.Reader` (as in *Reading input line by line*). Each line comes
back as a *record*: a `[]string` with one string per field, quotes removed.

```go
r := csv.NewReader(strings.NewReader("hostname,region\nweb-1,eu-west-1\n"))
records, err := r.ReadAll()
// records => [][]string{{"hostname", "region"}, {"web-1", "eu-west-1"}}
```

`ReadAll` reads everything at once. You can also call `Read` in a loop, one
record at a time, until it returns `io.EOF`:

```go
for {
	record, err := r.Read()
	if err == io.EOF {
		break
	}
	if err != nil {
		return err
	}
	fmt.Println(record[0])
}
```

`ReadAll` is simplest for small files, so this exercise uses it. Empty input
gives zero records and no error.

By default, every record must have the same number of fields as the first
one. A row with too few or too many fields makes the reader return an error,
for example `record on line 3: wrong number of fields`. You can check for it
with `errors.Is(err, csv.ErrFieldCount)`.

### The header row

The first record is usually a *header*: the column names. Exports don't
always put the columns in the same order, so don't hard-code `record[2]`.
Instead, build a map from column name to position:

```go
header := records[0] // {"region", "hostname"}
col := make(map[string]int)
for i, name := range header {
	col[name] = i
}
// col => map[hostname:1 region:0]

_, ok := col["owner"] // ok is false: there is no owner column
```

Then read a field from any data row with `row[col["hostname"]]`.

### Writing

`csv.NewWriter` takes any `io.Writer`: something you can write bytes to, such
as a file, a network connection, or a `strings.Builder` in the tests.

```go
cw := csv.NewWriter(w)
cw.Write([]string{"web-1", "eu-west-1", "platform, sre"})
cw.Flush()
return cw.Error()
// output: web-1,eu-west-1,"platform, sre"
```

| Method | What it does |
|---|---|
| `Write(record)` | Adds one record. Quotes fields that need it and ends the line with `\n`. |
| `Flush()` | Pushes everything buffered so far to the real writer. |
| `Error()` | Returns any error from earlier `Write` or `Flush` calls. |

The writer collects output in a *buffer* and only sends it on when the buffer
fills up or you call `Flush`. Forget `Flush` and a small file comes out
**empty**. `Error()` after `Flush()` tells you whether anything went wrong, so
the usual ending is `cw.Flush()` then `return cw.Error()`.

## Your task

Your team keeps a CSV inventory of hosts. You need to load it, and write a
filtered copy for a single region.

Each host becomes a `Host{Name, Region, Owner}`. In the CSV, the columns are
called `hostname`, `region` and `owner`.

### 1. Read the inventory

Finish `ReadHosts`. It reads CSV from `r` and returns one `Host` per data row,
in file order.

- The first record is the header. Columns can be in any order, and extra
  columns (like `env`) are ignored.
- Input with no records at all returns the error `empty inventory`.
- If `hostname`, `region` or `owner` is missing from the header, return the
  error `missing column "<name>"`. Check them in that order and report the
  first one missing.
- If the csv reader returns an error, return it (wrapping it with `%w` is
  fine). The tests check `errors.Is(err, csv.ErrFieldCount)`.
- A header with no data rows returns no hosts and no error.

```go
ReadHosts(strings.NewReader("env,owner,hostname,region\nprod,payments,api-1,us-east-1\n"))
// => []Host{{Name: "api-1", Region: "us-east-1", Owner: "payments"}}, nil

ReadHosts(strings.NewReader("hostname,region\napi-1,us-east-1\n"))
// => nil, error `missing column "owner"`

ReadHosts(strings.NewReader(""))
// => nil, error `empty inventory`
```

You'll need `"errors"` and `"fmt"` for the error messages: add them to the
import block.

### 2. Write the inventory

Finish `WriteHosts`. It writes the header `hostname,region,owner`, then one
record per host in the same order. The flush and the error check are already
written for you.

```go
WriteHosts(w, []Host{
	{Name: "web-1", Region: "eu-west-1", Owner: "platform, sre"},
	{Name: "db-1", Region: "us-east-1", Owner: "data"},
})
```

writes:

```text
hostname,region,owner
web-1,eu-west-1,"platform, sre"
db-1,us-east-1,data
```

With no hosts, only the header line is written.

### 3. Filter by region

Finish `FilterCSV`. It reads an inventory from `r` and writes only the hosts
whose `Region` equals `region` to `w`, using the format from task 2. Reuse
`ReadHosts` and `WriteHosts`. If reading fails, return that error and write
nothing.

```go
in := "owner,hostname,region\npayments,api-1,us-east-1\ndata,db-1,eu-west-1\n"
FilterCSV(strings.NewReader(in), w, "us-east-1")
```

writes:

```text
hostname,region,owner
api-1,us-east-1,payments
```

If no host matches, only the header line is written.

## Learn more

- [Package encoding/csv](https://pkg.go.dev/encoding/csv)
- [Package io: Reader and Writer](https://pkg.go.dev/io)
- [RFC 4180: the CSV format](https://www.rfc-editor.org/rfc/rfc4180)
