## Task 1: Fix the usage line

Look at the signature `Percent(part, whole float64)`. Go matches arguments by
position, so the first value you pass becomes `part`. Which of `used` and
`total` is the part, and which is the whole?

### Closer

`used` is the part and `total` is the whole, so `used` goes first. Swap the two
arguments inside the `Percent(...)` call and leave the rest of the line alone.

### Answer

```go
return fmt.Sprintf("%s: %.1f%% CPU", service, Percent(used, total))
```

## Task 2: Write the tag helper

An empty string `""` is what a `string` holds when nothing was given. Check
whether `tag` is empty, and pick what to return in each case.

### Closer

Use an `if` that compares with `==`: `if tag == "" { ... }`. Inside, return
the fixed text `"latest"` (in quotes). After the `if`, return the parameter
`tag` (no quotes).

### Answer

```go
if tag == "" {
	return "latest"
}
return tag
```

## Task 3: Build the image reference

The result is three pieces joined by `/` and `:`. Your `TagOrLatest` from task
2 already picks the right tag, so call it instead of using `tag` directly.

### Closer

Fixed text goes in quotes (`"/"` and `":"`); parameters and function calls
don't. Join every piece with `+`: `registry + "/" + name + ...`.

### Answer

```go
return registry + "/" + name + ":" + TagOrLatest(tag)
```
