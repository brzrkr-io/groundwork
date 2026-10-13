## Task 1: Parse the inventory

Make a `json.Decoder` that reads from `r`, tell it to reject unknown keys, and
decode into a `[]Host`. Any error gets the `parse inventory: ` prefix.

### Closer

`dec.DisallowUnknownFields()` turns a typo like `"regoin"` into an error. Wrap
with `%w`: `fmt.Errorf("parse inventory: %w", err)`.

### Answer

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()
var hosts []Host
if err := dec.Decode(&hosts); err != nil {
	return nil, fmt.Errorf("parse inventory: %w", err)
}
return hosts, nil
```

## Task 2: Validate every host

Follow the region check the starter gives you: one `if` per rule, each
appending an error to `errs`, in the order of the table. For duplicates you
need to remember where each name was first seen.

### Closer

Use `slices.Contains(Envs, h.Env)` for the env rule and the error from
`netip.ParseAddr(h.Address)` for the address rule. For duplicates, keep
`firstSeen := map[string]int{}` from name to host number, and only check it
when `h.Name != ""`.

### Answer

```go
if h.Name == "" {
	errs = append(errs, fmt.Errorf("%s: name is required", where))
}
if !slices.Contains(Envs, h.Env) {
	errs = append(errs, fmt.Errorf("%s: env %q must be one of %s", where, h.Env, strings.Join(Envs, ", ")))
}
// region check (already written)
if _, err := netip.ParseAddr(h.Address); err != nil {
	errs = append(errs, fmt.Errorf("%s: address %q is not a valid IP address", where, h.Address))
}
if h.Name != "" {
	if first, ok := firstSeen[h.Name]; ok {
		errs = append(errs, fmt.Errorf("%s: duplicate name, first seen at host %d", where, first))
	} else {
		firstSeen[h.Name] = n
	}
}
```

Declare `firstSeen` before the loop, and `return errors.Join(errs...)` after it.

## Task 3: Match hosts against a filter

Return `false` as soon as one part of the filter fails, and `true` at the end.
An empty `f.Env` or `f.Region` means "don't care".

### Closer

Each field check looks like `if f.Env != "" && h.Env != f.Env { return false }`.
For tags, loop over `f.Tags` and use `slices.Contains(h.Tags, tag)`.

### Answer

```go
if f.Env != "" && h.Env != f.Env {
	return false
}
if f.Region != "" && h.Region != f.Region {
	return false
}
for _, tag := range f.Tags {
	if !slices.Contains(h.Tags, tag) {
		return false
	}
}
return true
```

## Task 4: Print a table

Write the header and every row into a `tabwriter`, with `\t` between cells and
`\n` at the end of each row. Nothing appears until you call `Flush`.

### Closer

Start with `tags := "-"` and replace it with `strings.Join(h.Tags, ",")` when
the host has tags. A row is
`fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", ...)`.

### Answer

```go
tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
fmt.Fprintln(tw, "NAME\tENV\tREGION\tADDRESS\tTAGS")
for _, h := range hosts {
	tags := "-"
	if len(h.Tags) > 0 {
		tags = strings.Join(h.Tags, ",")
	}
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", h.Name, h.Env, h.Region, h.Address, tags)
}
return tw.Flush()
```

## Task 5: Print JSON

A `nil` slice encodes as `null`, so swap it for an empty slice first. Then let
an indenting `json.Encoder` do the rest.

### Closer

`hosts = []inventory.Host{}` when `hosts == nil`. Then
`enc := json.NewEncoder(w)` and `enc.SetIndent("", "  ")`.

### Answer

```go
if hosts == nil {
	hosts = []inventory.Host{}
}
enc := json.NewEncoder(w)
enc.SetIndent("", "  ")
return enc.Encode(hosts)
```

## Task 6: Parse the flags

Define `-env` and `-region` so they write straight into `opts.filter`, and
`-tag` with `fs.Func` so each use appends. After the leftover-argument check,
reject an unknown env with `usageError`.

### Closer

`fs.StringVar(&opts.filter.Env, "env", "", "...")` stores into the struct
field. The env check is
`env != "" && !slices.Contains(inventory.Envs, env)`.

### Answer

```go
fs.StringVar(&opts.filter.Env, "env", "", "only hosts in this `env`")
fs.StringVar(&opts.filter.Region, "region", "", "only hosts in this `region`")
fs.Func("tag", "only hosts with this `tag`", func(tag string) error {
	opts.filter.Tags = append(opts.filter.Tags, tag)
	return nil
})
// after fs.Parse and the leftover-argument check:
if env := opts.filter.Env; env != "" && !slices.Contains(inventory.Envs, env) {
	return options{}, usageError(fs, "-env must be one of %s, got %q", strings.Join(inventory.Envs, ", "), env)
}
```

## Task 7: Wire up Run

Follow steps 3 to 5 of the README in order: validate, filter and sort, print.
Every failure prints `inventory: ...` to `stderr` and returns `1`.

### Closer

The validate message is
`fmt.Fprintf(stderr, "inventory: %s is invalid:\n%v\n", opts.file, err)`. Sort
with `slices.SortFunc(hosts, func(a, b inventory.Host) int { return cmp.Compare(a.Name, b.Name) })`.

### Answer

```go
if err := inventory.Validate(hosts); err != nil {
	fmt.Fprintf(stderr, "inventory: %s is invalid:\n%v\n", opts.file, err)
	return 1
}
hosts = opts.filter.Apply(hosts)
slices.SortFunc(hosts, func(a, b inventory.Host) int {
	return cmp.Compare(a.Name, b.Name)
})
if opts.json {
	err = report.WriteJSON(stdout, hosts)
} else {
	err = report.WriteTable(stdout, hosts)
}
if err != nil {
	fmt.Fprintf(stderr, "inventory: %v\n", err)
	return 1
}
return 0
```
