## Task 1: Add up a week of errors

You need to visit every one of the seven days and add its count to `total`. Indexes run from `0` to `len(daily)-1`.

### Closer

Loop over the indexes with `for i := range len(daily)` and read each day as `daily[i]`. Use `+=` to add to `total`.

### Answer

```go
for i := range len(daily) {
	total += daily[i]
}
```

The loop visits indexes `0` to `6`, so every day gets added once.

## Task 2: Name the zones of a region

Each zone name is `region` with one letter stuck on the end. You can build the whole array at once instead of filling it one element at a time.

### Closer

An array literal is the type followed by the values in braces: `[3]string{..., ..., ...}`. Join strings with `+`, like `region + "a"`.

### Answer

```go
return [3]string{region + "a", region + "b", region + "c"}
```

The literal lists the three names in order, so index `0` is the `a` zone.

## Task 3: Mark a zone as down

`zone` is an index into `status`. Change that one element, then hand the array back.

### Closer

Write to one element with `status[zone] = ...`. `status` is already a copy of the caller's array, so changing it is safe, but you must `return` it.

### Answer

```go
status[zone] = false
return status
```

The function got its own copy of the array, so the caller's array stays unchanged.
