# Bytes, runes and UTF-8: fix the byte bugs

Two helpers in your on-call tooling were tested only with English names and
ASCII passwords. Both break as soon as someone types `é` or `€`. Find the byte
assumption in each and fix it.

## Learn

### Recap

A string is UTF-8 bytes: `len(s)` and `s[i]` work on bytes, and one character
(a *rune*) can take 1 to 4 bytes. `utf8.RuneCountInString(s)` counts runes,
`for _, r := range s` loops over runes, and `[]rune(s)` gives you a slice you
can index and cut by character. `string(r)` and `string(runes)` turn runes back
into a string.

### Spot the byte assumption

Code like this looks fine and passes every ASCII test:

```go
first := string(name[0])         // name[0] is a byte, not a character
last4 := token[len(token)-4:]    // len and slicing count bytes
```

For `"élodie"`, `name[0]` is only the first of the two bytes of `é`, and
`string(...)` of that byte gives `"Ã"`. Whenever you see `len(s)`, `s[i]` or
`s[a:b]` on text a human typed, ask: "bytes or characters?"

## Your task

The pager shows each on-call engineer as a one-letter avatar, and the audit
log prints secrets with most characters hidden. Both functions use bytes where
they should use characters.

### 1. Fix the avatar letter

`Initial(name)` returns the first character of `name` in upper case (use
`strings.ToUpper`, which handles letters like `é` too). For an empty name it
returns `"?"`. The ASCII cases already work; the accented ones don't.

```go
Initial("alice")   // => "A"
Initial("élodie")  // => "É"
Initial("Øystein") // => "Ø"
Initial("")        // => "?"
```

### 2. Fix the secret mask

`MaskSecret(secret)` hides a secret for logging. It keeps the last 4
characters and replaces every character before them with one `*`. If the
secret has 4 characters or fewer, every character becomes `*`. The number of
`*` always matches the number of hidden **characters**, not bytes.

```go
MaskSecret("s3cr3t-t0ken") // => "********0ken"
MaskSecret("grüße2026")    // => "*****2026"
MaskSecret("geheim€€€€")   // => "******€€€€"
MaskSecret("abc")          // => "***"
MaskSecret("äöü")          // => "***"
MaskSecret("")             // => ""
```

## Learn more

- [Go blog: Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [Package unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [Package strings: Repeat](https://pkg.go.dev/strings#Repeat)
