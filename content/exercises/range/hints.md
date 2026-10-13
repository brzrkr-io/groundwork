## Task 1: element only

You need the values, not the positions. Use `for _, v := range ms` and add `v`
to `total`.

## Task 2: index and element

Use `for i, host := range hosts`. The latency for `host` is `ms[i]`. Replace
`slowest` and `worst` when this is the first host (`i == 0`) or when `ms[i]` is
strictly greater than `worst`. "Strictly greater" keeps the first host on a
tie.

## Task 3: filter in order

Range over `healthy` with `for i, ok := range healthy`. When `ok` is `false`
(`!ok`), append `hosts[i]` to `down`.

## Almost the answer

```go
for i, host := range hosts {
	if i == 0 || ms[i] > worst {
		slowest, worst = host, ms[i]
	}
}
```
