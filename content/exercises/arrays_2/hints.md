## Task 1: where does the loop start?

Indexes start at `0`. Look at the first part of the `for` line in
`TotalRequests`. Which hour never gets added? `for i := range len(hours)`
visits every index, from `0` to `23`.

## Task 2: index or count?

`best` holds an hour (an index). `hours[i]` is a request count. Comparing a
count with an hour makes no sense. Compare `hours[i]` with the count of the
best hour so far, which is `hours[best]`.

## Task 3: two separate arrays

`cleaned := hours` makes a full copy. After that line, `cleaned` and `hours`
are independent, so changing `hours` does nothing to the array you return.
Change `cleaned` instead.

## Almost the answer

```go
for i := range len(hours) { // task 1
```

```go
if hours[i] > hours[best] { // task 2
```

```go
cleaned[hour] = 0 // task 3
```
