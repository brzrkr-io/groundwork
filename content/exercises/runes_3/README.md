# Bytes, runes and UTF-8: DNS-safe service names

People name services "Payments API" or "Café Orders 🚀". Kubernetes and DNS
want `payments-api`. In this exercise you turn any display name into a safe
name, rune by rune.

## Learn

### Recap

`for _, r := range s` gives you each rune (character) of a string, and runes
are numbers, so `'a' <= r && r <= 'z'` checks for a lower-case ASCII letter.
`unicode.IsLetter` is the wrong check here, because it also accepts `é`. Build
the output with a `strings.Builder`, as in *The strings package*.

### Writing runes to a Builder

`b.WriteString(s)` adds a string. `b.WriteRune(r)` adds one rune, so you don't
need `string(r)` first:

```go
var b strings.Builder
b.WriteRune('a')
b.WriteRune('-')
b.WriteString("api")
b.String() // "a-api"
b.Len()    // 5 (bytes written so far)
```

### Cutting ASCII is safe

Slicing with `s[:n]` counts bytes. That's dangerous for human text, but once a
string contains only ASCII characters, one byte is one character, and `s[:n]`
is exactly the first `n` characters.

## Your task

Your platform's "create service" form takes a display name, and the tooling
needs a DNS-safe name for the Kubernetes Service.

### 1. Slugify a display name

`Slug(name)` turns a display name into a lower-case ASCII name:

- ASCII letters are kept and turned into lower case (`A`–`Z` become `a`–`z`).
- ASCII digits `0`–`9` are kept.
- Every other character is a separator: spaces, punctuation, `_`, `-`,
  emoji, and non-ASCII letters like `é`.
- Each run of one or more separators between kept characters becomes a single
  `-`.
- Separators at the start or end produce nothing, so the result never starts
  or ends with `-`.
- If nothing is kept, the result is `""`.

```go
Slug("Payments API")           // => "payments-api"
Slug("orders")                 // => "orders"
Slug("  Billing -- Worker!! ") // => "billing-worker"
Slug("My_Service.v2")          // => "my-service-v2"
Slug("eu-west-1 Gateway")      // => "eu-west-1-gateway"
Slug("🚀 Launch Service 🚀")    // => "launch-service"
Slug("Café Orders")            // => "caf-orders"
Slug("***")                    // => ""
```

Yes, `"Café"` becomes `"caf"`: `é` is not ASCII, so it counts as a separator.
Add `"strings"` to the imports to use `strings.Builder`.

### 2. Make a valid DNS label

A DNS label has at most 63 characters. `DNSLabel(name)` returns `Slug(name)`,
with three extra rules, applied in this order:

1. If the slug is longer than 63 characters, keep only the first 63.
2. If that cut leaves a `-` at the end, remove it (there is at most one,
   because `Slug` never produces two in a row).
3. If the result is empty, return `"unnamed"`.

```go
DNSLabel("Payments API")                      // => "payments-api"
DNSLabel("!!!")                               // => "unnamed"
DNSLabel("日本語")                             // => "unnamed"
DNSLabel(strings.Repeat("a", 70))             // => 63 "a"s
DNSLabel(strings.Repeat("a", 62) + " backup") // => 62 "a"s (the cut left "aaa…a-")
```

## Learn more

- [Kubernetes: DNS label names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names)
- [Package strings: Builder](https://pkg.go.dev/strings#Builder)
- [Go blog: Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
