## Task 1: two trims and a cut

Call `strings.TrimPrefix` once for `"https://"` and once for `"http://"`.
Only one of them will match, and the other does nothing. Then
`strings.Cut(url, "/")` gives you the part before the first `"/"`.

## Task 2: split, then cut

`strings.SplitSeq(s, ",")` gives you the pairs, one at a time. Range over
them with `for pair := range strings.SplitSeq(s, ",")`, and for each pair
call `strings.Cut(pair, "=")`. If the third result is `false`, use `continue`
to skip it.

## Task 3: space first, then maybe quotes

Inside the loop, always `b.WriteString(" ")` first. Then, if
`strings.Contains(arg, " ")`, write `"\""`, the argument and another `"\""`;
otherwise just write the argument.

## Almost the answer for task 2

```go
for pair := range strings.SplitSeq(s, ",") {
	key, value, found := strings.Cut(pair, "=")
	if !found {
		continue
	}
	labels[strings.TrimSpace(key)] = strings.TrimSpace(value)
}
```
