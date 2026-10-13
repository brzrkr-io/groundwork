## Task 1: one argument at a time

Start a `strings.Builder` and write `bin` into it. Loop over `args`, and for
each one write a space and then the argument. Before writing it, wrap it in
quotes if it is `""` or if `strings.Contains(a, " ")` is true. A backtick
string such as `` `"` `` is an easy way to write a double quote. Return
`b.String()` at the end.

## Task 2: build one slice

You can't write `Command("kubectl", "-n", namespace, args...)`. Build a single
slice instead: start with `var all []string`, append `"-n"` and `namespace`
only when the namespace is not empty, then append `args...`. Finally call
`Command("kubectl", all...)`.

## Task 3: like JoinPath

This is the same pattern as `JoinPath` in *Variadic functions*, with
`" && "` instead of `"/"`:

```go
if c == "" {
	continue
}
if script != "" {
	script += " && "
}
script += c
```
