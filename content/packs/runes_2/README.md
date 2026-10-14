# Bytes, runes and UTF-8: fix the byte bugs

Two helpers in your on-call tooling were tested only with English names and
ASCII passwords. Both break as soon as someone types `é` or `€`. Find the byte
assumption in each and fix it.

**New here:** nothing; practice finding byte bugs.

## Learn

### Recap

- A string is UTF-8 *bytes*. One character (a *rune*) takes 1 to 4 bytes.
- `len(s)`, `s[i]` and `s[a:b]` all work on **bytes**.
- `[]rune(s)` gives a slice with one element per character; `len`, `[i]` and
  `[a:b]` on it count characters.
- `string(r)` and `string(runes)` turn runes back into a string.

### Spot the byte assumption

This looks fine and passes every ASCII test:

```go
grade := string(score[0])        // score[0] is a byte, not a character
suffix := file[len(file)-3:]     // len and slicing count bytes
```

For `"élodie"`, the first byte is only half of `é`; `string` of it gives
`"Ã"`. Whenever you see `len(s)`, `s[i]` or `s[a:b]` on text a human typed,
ask: "bytes or characters?"

> **Why is it written like this?** Go requires the conversion; it will never
> turn a string into runes for you. Converting once into a variable (say
> `runes`) and reusing it is the author's choice; it saves converting again on
> every line.

### `strings.Repeat`

`strings.Repeat(s, count)` returns `s` written `count` times:

```go
strings.Repeat("=", 3) // "==="
strings.Repeat("=", 0) // ""
```

## Your task

The pager shows each on-call engineer as a one-letter avatar, and the audit
log prints secrets with most characters hidden. Both functions use bytes where
they should use characters.

### 1. Fix the avatar letter

**Goal:** make `Initial(name)` return the first **character** of `name` in upper case, or `"?"` if `name` is empty.

Upper-case it with `strings.ToUpper` (already used; it handles `é` too). ASCII
names already work; accented ones don't.

```go
Initial("alice")   // => "A"
Initial("Bob")     // => "B"
Initial("élodie")  // => "É"
Initial("Øystein") // => "Ø"
Initial("żaneta")  // => "Ż"
Initial("")        // => "?"
```

**Think it through:**
- Which expression in the starter grabs a byte instead of a character?
- What do you convert `name` to so that `[0]` is a whole character?
- Keep the empty check first: why would `[0]` fail on an empty name?

### 2. Fix the secret mask

**Goal:** make `MaskSecret(secret)` keep the last 4 characters and replace each earlier character with one `*`.

- If the secret has 4 characters or fewer, every character becomes `*`.
- The number of `*` matches the number of hidden **characters**, not bytes.

```go
MaskSecret("s3cr3t-t0ken")     // => "********0ken"
MaskSecret("grüße2026")        // => "*****2026"
MaskSecret("geheim€€€€")       // => "******€€€€"
MaskSecret("deploy-key-🔑🔑")  // => "*********y-🔑🔑"
MaskSecret("abcd")             // => "****"
MaskSecret("äöü")              // => "***"
MaskSecret("")                 // => ""
```

**Think it through:**
- Count every `len(secret)` and `secret[...]` in the starter. Each one is a
  byte bug.
- Convert once, then replace each of them with the character version.
- The last 4 are runes; how do you turn them back into a string to join with
  the stars?

## Learn more

- [Go blog: Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [Package unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [Package strings: Repeat](https://pkg.go.dev/strings#Repeat)
