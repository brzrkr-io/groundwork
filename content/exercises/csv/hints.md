## Task 1: Read the inventory

Read every record at once, check for errors and an empty file, then treat
`records[0]` as the header and `records[1:]` as the data rows. A map from
column name to position lets the columns come in any order.

### Closer

`csv.NewReader(r).ReadAll()` returns a `[][]string`. Build the map with
`for i, name := range records[0] { col[name] = i }`. Then, for each of
`"hostname"`, `"region"`, `"owner"`, check `_, ok := col[name]` and return
`fmt.Errorf("missing column %q", name)` when it's missing. Add `"errors"` and
`"fmt"` to the imports.

### Answer

```go
records, err := csv.NewReader(r).ReadAll()
if err != nil {
	return nil, err
}
if len(records) == 0 {
	return nil, errors.New("empty inventory")
}

col := make(map[string]int)
for i, name := range records[0] {
	col[name] = i
}
for _, name := range []string{"hostname", "region", "owner"} {
	if _, ok := col[name]; !ok {
		return nil, fmt.Errorf("missing column %q", name)
	}
}

var hosts []Host
for _, row := range records[1:] {
	hosts = append(hosts, Host{
		Name:   row[col["hostname"]],
		Region: row[col["region"]],
		Owner:  row[col["owner"]],
	})
}
return hosts, nil
```

## Task 2: Write the inventory

Each call to the csv writer's `Write` takes one record as a `[]string`.
Write the header first, then one record per host.

### Closer

`cw.Write([]string{"hostname", "region", "owner"})`, then loop over `hosts`
and write `[]string{h.Name, ...}`. The flush and error check are already
there, and the writer adds quotes when a field contains a comma.

### Answer

```go
cw.Write([]string{"hostname", "region", "owner"})
for _, h := range hosts {
	cw.Write([]string{h.Name, h.Region, h.Owner})
}
```

## Task 3: Filter by region

You already have a reader and a writer. Read the hosts, keep the matching
ones, and write those.

### Closer

Call `ReadHosts(r)` and return early on error, before writing anything.
Collect hosts with `h.Region == region` into a new slice and pass it to
`WriteHosts`.

### Answer

```go
hosts, err := ReadHosts(r)
if err != nil {
	return err
}

var keep []Host
for _, h := range hosts {
	if h.Region == region {
		keep = append(keep, h)
	}
}
return WriteHosts(w, keep)
```

An empty `keep` still writes the header, because `WriteHosts` always writes it.
