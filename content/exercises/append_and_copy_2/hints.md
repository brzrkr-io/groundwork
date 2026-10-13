## Look at the lengths

Both bugs are in the `make` line. Ask: how many elements does the slice hold
*before* the loop or `copy` runs?

## Task 1: append adds after what's there

`make([]string, len(hosts))` already holds `len(hosts)` empty strings, and
`append` adds after them. Start with length 0 instead.

## Task 2: copy only fills existing slots

`copy` copies the smaller of the two lengths. A destination of length 0 gets
nothing. Give `backup` the same length as `hosts`.

## Almost the answer

```go
out := make([]string, 0, len(hosts))
```

```go
backup := make([]string, len(hosts))
```
