## Work inside out

Do the tasks in order. `inventory` and `report` don't know about flags at
all, so you can finish and test them before touching `cli`. Each failing
test message names the function and shows what you returned next to what was
wanted.

## Tasks 1–3: inventory

```go
dec := json.NewDecoder(r)
dec.DisallowUnknownFields()
var hosts []Host
if err := dec.Decode(&hosts); err != nil {
	return nil, fmt.Errorf("parse inventory: %w", err)
}
return hosts, nil
```

For `Validate`, follow the region check the starter gives you: append
`fmt.Errorf("%s: ...", where)` to `errs` for each problem. For duplicates, keep a
`firstSeen := map[string]int{}` from name to host number. `Match` is three
`if` checks that `return false`, with `slices.Contains(h.Tags, tag)` inside a
loop over `f.Tags`.

## Tasks 4–5: report

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

For JSON, replace a `nil` slice with `[]inventory.Host{}` first, then
`enc := json.NewEncoder(w)`, `enc.SetIndent("", "  ")`, `return enc.Encode(hosts)`.

## Tasks 6–7: cli

```go
fs.StringVar(&opts.filter.Env, "env", "", "only hosts in this `env`")
fs.StringVar(&opts.filter.Region, "region", "", "only hosts in this `region`")
fs.Func("tag", "only hosts with this `tag`", func(tag string) error {
	opts.filter.Tags = append(opts.filter.Tags, tag)
	return nil
})
// after the leftover-argument check:
if env := opts.filter.Env; env != "" && !slices.Contains(inventory.Envs, env) {
	return options{}, usageError(fs, "-env must be one of %s, got %q", strings.Join(inventory.Envs, ", "), env)
}
```

`Run` follows the numbered steps in the README, top to bottom. After `load`:

```go
if err := inventory.Validate(hosts); err != nil {
	fmt.Fprintf(stderr, "inventory: %s is invalid:\n%v\n", opts.file, err)
	return 1
}
hosts = opts.filter.Apply(hosts)
// SortFunc with cmp.Compare, then WriteTable or WriteJSON.
```
