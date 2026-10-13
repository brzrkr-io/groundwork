# Bytes, runes and UTF-8

Service names, commit messages and alert texts are not always plain English.
A user called "Zoë" or a deploy message with a 🚀 breaks code that assumes one
letter is one byte. This exercise shows how Go stores text and how to count,
cut and check it safely.

## Learn

### A string is a row of bytes

Go stores a string as *bytes* in the UTF-8 encoding. Plain English letters,
digits and punctuation (the *ASCII* characters) take 1 byte each. Most other
characters take more:

| Text | Characters | `len` (bytes) |
|---|---|---|
| `"api"` | 3 | 3 |
| `"é"` | 1 | 2 |
| `"café"` | 4 | 5 |
| `"🚀"` | 1 | 4 |

So `len(s)` counts **bytes**, not characters. For ASCII the two numbers match,
which is why the difference is easy to miss.

### Runes

Go calls one character a *rune*. The type `rune` holds a number: the Unicode
*code point* of the character. A *rune literal* uses single quotes:

```go
var r rune = 'é' // single quotes: one rune
s := "é"         // double quotes: a string (2 bytes)
```

To count characters, use `utf8.RuneCountInString` from the `unicode/utf8`
package:

```go
import "unicode/utf8"

utf8.RuneCountInString("café") // 4
len("café")                    // 5
```

### Indexing gives bytes, range gives runes

`s[i]` gives you **one byte**, not one character. For `"café"`, `s[3]` is the
first half of `é`, which is meaningless on its own. Slicing with `s[:n]` also
counts bytes, so it can cut a character in half.

`for ... range` over a string, as you saw briefly in *Looping with range*,
decodes the UTF-8 for you. It gives the byte position where each rune starts,
and the rune itself:

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

### Converting to []rune and back

When you need to count or cut by characters, convert the string to a slice of
runes. Each element is one whole character, so indexing and slicing work the
way you'd expect. Convert back with `string(...)`:

```go
runes := []rune("🚀 shipped")
len(runes)             // 9 characters
string(runes[:1])      // "🚀"
string(runes[2:])      // "shipped"
```

### Checking what kind of character a rune is

The `unicode` package answers questions about a rune:

| Function | True for |
|---|---|
| `unicode.IsLetter(r)` | letters in **any** language: `a`, `Z`, `é`, `ж`, `中` |
| `unicode.IsDigit(r)` | digits in any script: `7`, and also `٣` (Arabic-Indic three) |
| `unicode.IsUpper(r)` | upper-case letters: `A`, `É` |
| `unicode.IsSpace(r)` | spaces, tabs, newlines |

"Any language" is great for human text, but it is the wrong check when a
system only accepts ASCII. Then compare the rune to ranges yourself. Runes are
numbers, and `'a'` to `'z'` sit next to each other in order:

```go
isLower := 'a' <= r && r <= 'z'
isDigit := '0' <= r && r <= '9'
```

## Your task

Your team's deploy bot posts messages to chat and adds labels to Kubernetes
objects. Its text helpers only work for plain English. Fix them so they handle
any text.

### 1. Count characters

`CharCount(s)` returns the number of characters (runes) in `s`, not the number
of bytes. Add `"unicode/utf8"` to the imports.

```go
CharCount("deploy")   // => 6
CharCount("café")     // => 4
CharCount("🚀 done")  // => 6
CharCount("")         // => 0
```

### 2. Truncate a chat message

The chat preview only has room for a few characters. `Truncate(msg, n)`
returns `msg` unchanged if it has `n` characters or fewer. Otherwise it returns
the first `n` characters followed by `"..."`. Count characters, never bytes, so
an emoji or an accented letter is never cut in half. `n` is never negative.

```go
Truncate("deploy finished", 6) // => "deploy..."
Truncate("ok", 10)             // => "ok"
Truncate("canary", 6)          // => "canary" (exactly 6, unchanged)
Truncate("café crashed", 4)    // => "café..."
Truncate("🚀🚀🚀 launch", 2)    // => "🚀🚀..."
```

### 3. Validate a label value

Kubernetes *labels* are key/value tags like `env=production`. `ValidLabel(s)`
reports whether `s` is a valid label value under these rules:

- It has 1 to 63 characters. (Kubernetes itself also accepts an empty value,
  but your team requires one.)
- Every character is an ASCII letter (`a`–`z`, `A`–`Z`), an ASCII digit
  (`0`–`9`), `-`, `_` or `.`.
- The first and the last character are ASCII letters or digits.

First write a small helper function, `isAlphaNum(r rune) bool`, anywhere in
`exercise.go`: it returns true only for `a`–`z`, `A`–`Z` and `0`–`9`. Don't
use `unicode.IsLetter` or `unicode.IsDigit`: they accept `é` and `٣`, which
Kubernetes rejects. Then use `isAlphaNum` in `ValidLabel` for the character
checks.

```go
ValidLabel("production")    // => true
ValidLabel("v1.2.3")        // => true
ValidLabel("Api_gateway-2") // => true
ValidLabel("")              // => false (empty)
ValidLabel("-canary")       // => false (starts with '-')
ValidLabel("canary.")       // => false (ends with '.')
ValidLabel("eu west")       // => false (space)
ValidLabel("café")          // => false (é is not ASCII)
```

A value of 63 `a`s is valid; 64 `a`s is too long.

## Learn more

- [Go blog: Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [Go by Example: Strings and Runes](https://gobyexample.com/strings-and-runes)
- [Package unicode](https://pkg.go.dev/unicode)
- [Kubernetes: Labels, syntax and character set](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set)
