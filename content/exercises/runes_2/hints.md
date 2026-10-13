## Find the byte operations

Look for `name[0]`, `len(secret)` and `secret[...]`. Each one works on bytes.
Replace them with operations on runes.

## Convert once, then work on runes

`[]rune(s)` gives one element per character. After that, `len`, indexing and
slicing all count characters:

```go
runes := []rune(secret)
len(runes)          // number of characters
string(runes[2:])   // characters from position 2 to the end
```

## Almost the answer

```go
first := []rune(name)[0]
return strings.ToUpper(string(first))
```

For `MaskSecret`, use `len(runes)` everywhere the starter uses `len(secret)`,
and `string(runes[hidden:])` instead of `secret[len(secret)-4:]`, where
`hidden := len(runes) - 4`.
