## Task 1: loop over the indexes

`len(daily)` is 7. Loop with `for i := range len(daily)` and add `daily[i]` to
`total` each round.

## Task 2: an array literal

You can build the whole array in one `return` with an array literal. Each
element is `region` joined with one letter using `+`.

## Task 3: change one element

Write to the element at index `zone` with `status[zone] = ...`, then return
`status`. The caller keeps its own copy, so nothing else is needed.

## Almost the answer

```go
for i := range len(daily) {
	total += daily[i]
}
```

```go
return [3]string{region + "a", region + "b", region + "c"}
```

```go
status[zone] = false
return status
```
