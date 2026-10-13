## Task 1: Fix the host names

Run the tests and compare got with want. Look at the first number in the
output, and at the very end of the string. There are two bugs.

### Closer

`for i := range count` gives `i` = 0, 1, 2, ..., but hosts start at 1, so
use `i+1` in the name. Then write the comma *before* each name except the
first (`if i > 0 { ... }`), instead of after every name.

### Answer

```go
for i := range count {
	if i > 0 {
		b.WriteString(",")
	}
	name := prefix + "-" + strconv.Itoa(i+1)
	b.WriteString(name)
}
```

## Task 2: Skip the reserved port

The loop stops at the reserved port, so every port after it is lost. You
want to skip just that one round and keep going.

### Closer

`break` leaves the loop for good. `continue` skips only the rest of this
round. Swap one for the other.

### Answer

```go
if port == reserved {
	continue
}
```

## Task 3: Count the batches

Compare the test cases: which one is wrong, the one that divides evenly
(`9, 3`) or the one with leftovers (`10, 3`)?

### Closer

The loop runs while `remaining >= batchSize`, so a leftover smaller than a
full batch is never counted. Keep looping as long as any items remain.

### Answer

```go
for remaining > 0 {
	remaining -= batchSize
	batches++
}
```
