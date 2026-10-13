## Which variable is which?

With one loop variable, `range` gives you the index. With two, the second one
is a copy of the element.

## Task 1: you're adding positions

`for minutes := range outages` makes `minutes` 0, 1, 2, …, the indexes. Put
`_, ` in front so `minutes` is the element instead.

## Task 2: write through the index

`v = limit` only changes the copy. Keep the index (`for i, v := range ms`) and
assign `ms[i] = limit`.

## Almost the answer

```go
for _, minutes := range outages {
	total += minutes
}
```

```go
for i, v := range ms {
	if v > limit {
		ms[i] = limit
	}
}
```
