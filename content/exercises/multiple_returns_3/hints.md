## Task 1: Cut, then check each case

```go
name, tag, found := strings.Cut(ref, ":")
```

Check `name == ""` first. Then, if `!found`, the tag is `"latest"`.
Otherwise, an empty `tag` is an error.

## Task 2: find the unit, then the number

Use `strings.CutSuffix` three times, once per unit, and remember how many
seconds that unit is worth (1, 60, 3600). Then convert the rest with
`strconv.Atoi`.

## Task 2: a shape that works

```go
multiplier := 0
num := ""
if rest, found := strings.CutSuffix(ttl, "s"); found {
	num, multiplier = rest, 1
} else if rest, found := strings.CutSuffix(ttl, "m"); found {
	num, multiplier = rest, 60
} else if rest, found := strings.CutSuffix(ttl, "h"); found {
	num, multiplier = rest, 3600
} else {
	return 0, fmt.Errorf("unknown unit in %q", ttl)
}
```

Then call `strconv.Atoi(num)`, check the error, reject `n < 1`, and return
`n * multiplier, nil`.
