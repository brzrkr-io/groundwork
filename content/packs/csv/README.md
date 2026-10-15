# CSV files: a host inventory

Spreadsheets, cloud billing exports, asset inventories and monitoring tools
all speak *CSV* (comma-separated values). Go's `encoding/csv` package reads
and writes it correctly, so you can turn an inventory export into Go structs
and back again.

**New here:** `encoding/csv`, `csv.NewReader`, `ReadAll`, `csv.NewWriter`, `Flush`.

## Learn

### Why not `strings.Split`?

A CSV line looks easy to split on commas, until a field contains a comma:

```text
hostname,region,owner
web-1,eu-west-1,"platform, sre"
```

`strings.Split` would give four pieces for the second line. CSV wraps a
field in double quotes when it contains a comma, a quote or a newline, and
writes a quote inside a field as two quotes (`"say ""hi"""`). The `csv`
package handles all of that for you.

### Reading

`csv.NewReader` takes any `io.Reader` (as in *Reading input line by line*).
Each line comes back as a *record*: a `[]string`, one string per field,
quotes removed.

```go
r := csv.NewReader(strings.NewReader("hostname,region\nweb-1,eu-west-1\n"))
records, err := r.ReadAll()
// records => [][]string{{"hostname", "region"}, {"web-1", "eu-west-1"}}
```

`ReadAll` returns a `[][]string`: a slice of records. `records[0]` is the
first line, `records[1:]` is everything after it. Empty input gives zero
records and no error. (You can also call `Read` in a loop until `io.EOF`;
*CSV files: fix the TSV DNS export* does that.)

By default, every record must have as many fields as the first one. A row
with too few or too many makes the reader return an error such as
`record on line 3: wrong number of fields`, which matches
`errors.Is(err, csv.ErrFieldCount)`.

### The header row

The first record is usually a *header*: the column names. Exports don't
always keep the same column order, so don't hard-code `row[2]`. Build a map
from column name to position instead:

```go
header := []string{"port", "service"}
pos := make(map[string]int)
for i, name := range header {
	pos[name] = i
}
// pos => map[port:0 service:1]

_, ok := pos["protocol"] // ok is false: no such column
```

Then read a field from any data row with `row[pos["service"]]`.

> **Why is it written like this?** Go requires `make` (or a `{}` literal)
> before you can store keys in a map. `for i, name := range` gives the index
> first and the value second; that order is fixed by Go. The names `pos`,
> `i` and `name` are the author's choice.

### Writing

`csv.NewWriter` takes any `io.Writer`: a file, a network connection, or a
`strings.Builder` in the tests.

```go
cw := csv.NewWriter(os.Stdout)
cw.Write([]string{"cache-1", "6379", "redis, primary"})
cw.Flush()
// output: cache-1,6379,"redis, primary"
```

| Method | What it does |
|---|---|
| `Write(record)` | Adds one record. Quotes fields that need it and ends the line with `\n`. |
| `Flush()` | Pushes everything buffered so far to the real writer. |
| `Error()` | Returns any error from earlier `Write` or `Flush` calls. |

The writer collects output in a *buffer* and only sends it on when the
buffer fills up or you call `Flush`. Forget `Flush` and a small file comes
out **empty**.

> **Why is it written like this?** `Write` takes a `[]string`, so Go requires
> you to build a slice even for fixed text: `[]string{"a", "b"}`. Checking
> `Error()` once after `Flush()`, instead of after every `Write`, is how the
> package is designed.

## Your task

Your team keeps a CSV inventory of hosts. You need to load it, and write a
filtered copy for a single region.

Each host becomes a `Host{Name, Region, Owner}`. In the CSV, those columns
are called `hostname`, `region` and `owner`. Note: the Go field is `Name`,
but the column is `hostname`.

### 1. Read the inventory

**Goal:** `ReadHosts(r)` reads CSV from `r` and returns one `Host` per data
row, in file order.

- The first record is the header. Columns can be in any order; extra columns
  (like `env`) are ignored.
- No records at all → error with the text `empty inventory`.
- `hostname`, `region` or `owner` missing from the header → error
  `missing column "<name>"`, with the name in double quotes. Check them in
  that order and report the first one missing.
- If the csv reader returns an error, return it (wrapping with `%w` is
  fine). The tests check `errors.Is(err, csv.ErrFieldCount)`.
- A header with no data rows → no hosts (`nil`) and no error.
- Add `"errors"` and `"fmt"` to the import block.

```go
ReadHosts(strings.NewReader("env,owner,hostname,region\nprod,payments,api-1,us-east-1\n"))
// => []Host{{Name: "api-1", Region: "us-east-1", Owner: "payments"}}, nil

ReadHosts(strings.NewReader("hostname,region\napi-1,us-east-1\n"))
// => nil, error `missing column "owner"`

ReadHosts(strings.NewReader(""))
// => nil, error `empty inventory`
```

**Think it through:**
- Order matters: read, then check for zero records, then build the column map, then check the three names.
- Which `fmt` verb puts double quotes around a string for you?
- Data rows are every record except the first. Which slice expression skips one?

### 2. Write the inventory

**Goal:** `WriteHosts(w, hosts)` writes the header `hostname,region,owner`,
then one record per host in the same order.

The flush and the error check are already written for you.

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

**Think it through:**
- The header is fixed text; each host row comes from the `Host` fields. In which order?
- The quotes around `platform, sre` are added by the writer. Don't add them yourself.

### 3. Filter by region

**Goal:** `FilterCSV(r, w, region)` reads an inventory from `r` and writes to
`w` only the hosts whose `Region` equals `region`, in the task 2 format.

Reuse `ReadHosts` and `WriteHosts`. If reading fails, return that error and
write nothing.

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

**Think it through:**
- Where do you check the read error so that nothing reaches `w`?
- Build a new slice of the matching hosts, then hand it to one function.

## Learn more

- [Package encoding/csv](https://pkg.go.dev/encoding/csv)
- [Package io: Reader and Writer](https://pkg.go.dev/io)
- [RFC 4180: the CSV format](https://www.rfc-editor.org/rfc/rfc4180)
