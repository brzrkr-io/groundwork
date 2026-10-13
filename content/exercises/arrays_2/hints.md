## Task 1: Count every hour

Indexes start at `0`. Look at the first part of the `for` line in `TotalRequests`: which hour never gets added?

### Closer

The loop starts at `i := 1`, so hour `0` (midnight) is skipped. A `range` loop over `len(hours)` visits every index from `0` to `23`.

### Answer

```go
for i := range len(hours) {
	total += hours[i]
}
```

`range len(hours)` starts at `0`, so midnight is counted too.

## Task 2: Find the peak hour

`best` holds an hour (an index), but `hours[i]` is a request count. The `if` compares a count with an hour, which makes no sense.

### Closer

Compare `hours[i]` with the request count of the best hour so far. You get that count by indexing: `hours[best]`.

### Answer

```go
if hours[i] > hours[best] {
	best = i
}
```

Strictly greater (`>`) means a later hour with the same count never replaces an earlier one.

## Task 3: Clear one bad hour

`cleaned := hours` makes a full copy. After that line the two arrays are independent. Which one does the function return, and which one does it change?

### Closer

The function returns `cleaned`, but the buggy line changes `hours`. Set the element in the array you return.

### Answer

```go
cleaned[hour] = 0
```

Changing `cleaned` changes the array that is returned; `hours` is a separate copy.
