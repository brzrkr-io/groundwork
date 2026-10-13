# Bytes, runes and UTF-8

Names like "Zoë" or a deploy message with a 🚀 break code that assumes one
letter is one byte. Here you learn how Go stores text, and how to count, cut
and check it safely.

**New here:** bytes vs runes, `rune`, `'x'` rune literals, `unicode/utf8`, `[]rune(s)`, comparing runes.

## Learn

### A string is a row of bytes

Go stores a string as *bytes*, in the *UTF-8* encoding. Plain English letters,
digits and punctuation (*ASCII*) take 1 byte each. Most other characters take
2 to 4 bytes:

| Text | Characters | `len` (bytes) |
|---|---|---|
| `"api"` | 3 | 3 |
| `"é"` | 1 | 2 |
| `"café"` | 4 | 5 |
| `"🚀"` | 1 | 4 |

So `len(s)` counts **bytes**, not characters. For ASCII the numbers match,
which is why the bug is easy to miss.

### Runes

Go calls one character a *rune*. A `rune` is a number: the character's
Unicode *code point*.

```go
var r rune = 'ж' // single quotes: one rune
s := "ж"         // double quotes: a string (2 bytes)
```

> **Why is it written like this?** Go requires single quotes for a rune and
> double quotes for a string; `'ab'` does not compile. The name `r` for a rune
> is convention.

### Counting runes: `unicode/utf8`

The `unicode/utf8` package decodes UTF-8 for you. To use it, add it to the
imports at the top of the file:

```go
import "unicode/utf8"

utf8.RuneLen('ж')            // 2: bytes this rune needs
utf8.ValidString("zoë")      // true
```

Browse [the package](https://pkg.go.dev/unicode/utf8): one of its functions
counts the runes in a string.

> **Why is it written like this?** Go requires every package you use to be
> imported, and an unused import is a compile error. You call its functions as
> `packagename.Function`, using the last part of the import path (`utf8`).

### Indexing gives bytes, range gives runes

`s[i]` gives **one byte**. For `"café"`, `s[3]` is the first half of `é`,
meaningless on its own. `s[:n]` also counts bytes, so it can cut a character
in half.

`for ... range` over a string (as in *Looping with range*) decodes UTF-8. It
gives the byte position where each rune starts, and the rune:

```go
for i, r := range "zoë!" {
	fmt.Println(i, string(r))
}
// 0 z
// 1 o
// 2 ë
// 4 !   (ë took bytes 2 and 3)
```

`string(r)` turns one rune back into a string.

### Converting to `[]rune` and back

To count or cut by characters, convert the string to a *slice of runes*. Each
element is one whole character. Convert back with `string(...)`:

```go
runes := []rune("🚀 shipped")
len(runes)        // 9 characters
string(runes[2:]) // "shipped"
```

> **Why is it written like this?** `[]rune(s)` and `string(runes)` are Go's
> required *conversion* syntax: the type name used like a function.

### What kind of character is it?

The `unicode` package has `unicode.IsLetter(r)` and `unicode.IsDigit(r)`. They
accept letters and digits from **every** language: `é`, `ж`, and `٣`
(Arabic-Indic three). That's wrong when a system only accepts ASCII.

Then compare ranges yourself. Runes are numbers, and `'a'` to `'z'` sit next
to each other in order:

```go
isVowelRange := 'a' <= r && r <= 'e'   // true for a, b, c, d, e
isSmallDigit := '0' <= r && r <= '3'   // true for 0, 1, 2, 3
```

`&&` means "and", `||` means "or", and `!` means "not".

### A helper function

A *helper* is a small function you write yourself, outside the others, and
call from them. Lower-case names stay private to the package:

```go
func isHex(r rune) bool {
	return ('0' <= r && r <= '9') || ('a' <= r && r <= 'f')
}
```

> **Why is it written like this?** Go requires `func`, the name, the
> parameters with their types in `( )`, then the result type, then `{`.
> The name and parameter names are your choice. Starting it lower-case is
> your choice too; it means "only this package uses it".

## Your task

Your deploy bot posts chat messages and adds labels to Kubernetes objects. Its
text helpers only work for plain English. Fix them so they handle any text.

### 1. Count characters

**Goal:** make `CharCount(s)` return the number of characters (runes) in `s`, not bytes.

```go
CharCount("deploy")  // => 6
CharCount("café")    // => 4
CharCount("🚀 done") // => 6
CharCount("сервер")  // => 6
CharCount("")        // => 0
```

**Think it through:**
- The starter returns `len(s)`. What does `len` count?
- Which package from *Learn* knows about runes? Remember to import it.

### 2. Truncate a chat message

**Goal:** make `Truncate(msg, n)` keep at most `n` characters, adding `"..."` when it cuts.

- If `msg` has `n` characters or fewer, return `msg` unchanged.
- Otherwise return the first `n` characters followed by `"..."`.
- Count characters, never bytes, so an emoji or accent is never cut in half.
  `n` is never negative.

```go
Truncate("deploy finished", 6) // => "deploy..."
Truncate("ok", 10)             // => "ok"
Truncate("canary", 6)          // => "canary"  (exactly 6: unchanged)
Truncate("café crashed", 4)    // => "café..."
Truncate("café", 4)            // => "café"
Truncate("🚀🚀🚀 launch", 2)    // => "🚀🚀..."
```

**Think it through:**
- What can you convert `msg` to so that `len` and `[:n]` count characters?
- After cutting, what type do you have, and how do you get a string back?
- `"..."` is fixed text; `msg` and `n` come from the parameters.

### 3. Validate a label value

**Goal:** make `ValidLabel(s)` return `true` only for a valid Kubernetes label value.

Kubernetes *labels* are tags like `env=production`; you check the value part.
Valid means all three:

- It has 1 to 63 characters. (Kubernetes allows empty values, but your team
  doesn't.)
- Every character is an ASCII letter (`a`–`z`, `A`–`Z`), an ASCII digit
  (`0`–`9`), `-`, `_` or `.`.
- The first and the last character are ASCII letters or digits.

First write a helper `isAlphaNum(r rune) bool` anywhere in `exercise.go`. It
returns true only for `a`–`z`, `A`–`Z` and `0`–`9`. Don't use
`unicode.IsLetter`/`IsDigit`: they accept `é` and `٣`, which Kubernetes
rejects. Then use `isAlphaNum` in `ValidLabel`.

```go
ValidLabel("production")    // => true
ValidLabel("v1.2.3")        // => true
ValidLabel("Api_gateway-2") // => true
ValidLabel("a")             // => true
ValidLabel("")              // => false (empty)
ValidLabel("-canary")       // => false (starts with '-')
ValidLabel(".hidden")       // => false (starts with '.')
ValidLabel("canary_")       // => false (ends with '_')
ValidLabel("eu west")       // => false (space)
ValidLabel("team/api")      // => false ('/')
ValidLabel("café")          // => false (é is not ASCII)
ValidLabel("v٣")            // => false (٣ is not ASCII)
ValidLabel("ship🚀it")      // => false (emoji)
```

63 `a`s is valid; 64 `a`s is too long.

**Think it through:**
- `isAlphaNum` needs three ranges joined with "or". Which operator is "or"?
- In `ValidLabel`, check the length first: why must that happen before you
  look at the first character?
- Which index is the last element of a slice of length `L`?

## Learn more

- [Go blog: Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [Go by Example: Strings and Runes](https://gobyexample.com/strings-and-runes)
- [Package unicode/utf8](https://pkg.go.dev/unicode/utf8)
- [Kubernetes: Labels, syntax and character set](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set)
