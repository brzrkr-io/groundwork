## Task 1: Fix the avatar letter

`name[0]` is the first **byte**, and `é` takes two bytes. You need the first
**character** instead.

### Closer

`[]rune(name)` gives one element per character, so `[]rune(name)[0]` is the
first character. Turn it back into a string with `string(...)` before calling
`strings.ToUpper`.

### Answer

```go
if name == "" {
	return "?"
}
first := []rune(name)[0]
return strings.ToUpper(string(first))
```

The empty check stays first, because an empty rune slice has no element `[0]`.

## Task 2: Fix the secret mask

Every `len(secret)` and `secret[...]` in the starter counts bytes. Convert the
secret once, then do the counting and cutting on characters.

### Closer

After `runes := []rune(secret)`, use `len(runes)` wherever the starter uses
`len(secret)`. The number of hidden characters is `hidden := len(runes) - 4`,
and `string(runes[hidden:])` is the last 4 characters.

### Answer

```go
runes := []rune(secret)
if len(runes) <= 4 {
	return strings.Repeat("*", len(runes))
}
hidden := len(runes) - 4
return strings.Repeat("*", hidden) + string(runes[hidden:])
```

Now the number of `*` matches the number of hidden characters, not bytes.
