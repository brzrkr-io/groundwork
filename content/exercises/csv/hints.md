## Task 1: start with ReadAll

`csv.NewReader(r).ReadAll()` gives you a `[][]string`. Check its error first,
then check `len(records) == 0` for the empty-inventory case. `records[0]` is
the header and `records[1:]` are the data rows.

## Task 1: the column map

Loop over the header with `for i, name := range records[0]` and store
`col[name] = i`. Then, for each of `"hostname"`, `"region"`, `"owner"`, use
`_, ok := col[name]` and return `fmt.Errorf("missing column %q", name)` when
`ok` is false.

## Tasks 2 and 3: write and filter

```go
cw.Write([]string{"hostname", "region", "owner"})
for _, h := range hosts {
	cw.Write([]string{h.Name, h.Region, h.Owner})
}
```

For `FilterCSV`, call `ReadHosts(r)`, return early on error, collect the hosts
with `h.Region == region` into a new slice, and `return WriteHosts(w, keep)`.

## Almost the answer (task 1)

```go
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
