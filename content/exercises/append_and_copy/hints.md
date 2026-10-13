## Assign what append returns

`append` gives back a new slice. Write `order = append(order, ...)`, not just
`append(order, ...)`. For all of `hosts` at once, use `hosts...`.

## Task 2: the loop

Use `for i := range len(hosts)` and, inside, append `hosts[i] + "." + domain`
to `out`. Remember to assign: `out = append(out, ...)`.

## Task 3: a real copy

`return hosts` hands back the same elements, so edits to `hosts` show up in
the "copy". Either make a new slice and copy into it:

```go
snap := make([]string, len(hosts))
copy(snap, hosts)
return snap
```

or add `import "slices"` at the top and `return slices.Clone(hosts)`.

## Almost the answer

```go
order = append(order, hosts...)
```

```go
for i := range len(hosts) {
	out = append(out, hosts[i]+"."+domain)
}
```
