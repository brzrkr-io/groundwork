## It's a slice

Inside each function, the `...` parameter is a normal slice: `values` is an
`[]int`, `parts` and `extra` are `[]string`. Use `len` and `range` on them as
you did in *Looping with range*.

## Task 1: start from the first value

Return `0` straight away when `len(values) == 0`. Otherwise keep the largest
value seen so far in a variable that starts as `values[0]`, then loop over the
rest and replace it whenever you find something bigger.

## Task 2: only add "/" between parts

Skip a part with `continue` when it is `""`. Before adding a part, add `"/"`
only if the result so far is not empty:

```go
if path != "" {
	path += "/"
}
path += p
```

## Task 3: append with ...

Start a new slice that holds `base`, then append every value of `extra` with
`...`:

```go
return append([]string{base}, extra...)
```
