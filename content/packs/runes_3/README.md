# Bytes, runes and UTF-8: DNS-safe service names

People name services "Payments API" or "Café Orders 🚀". Kubernetes and DNS
want `payments-api`. Here you turn any display name into a safe name, rune by
rune.

**New here:** `b.WriteRune`, `b.Len`, rune arithmetic, a "remember for later" flag.

## Learn

### Recap

`for _, r := range s` gives you each rune of a string. Runes are numbers, so
`'a' <= r && r <= 'z'` checks for a lower-case ASCII letter.
`unicode.IsLetter` is the wrong check here: it also accepts `é`. Build output
with a `strings.Builder`, as in *The strings package*.

### Writing runes to a Builder

`b.WriteString(s)` adds a string. `b.WriteRune(r)` adds one rune, no
`string(r)` needed. `b.Len()` tells you how many bytes have been written so far:

```go
var b strings.Builder
b.Len()           // 0: nothing yet
b.WriteRune('x')
b.WriteRune('/')
b.WriteString("y")
b.String()        // "x/y"
b.Len()           // 3
```

> **Why is it written like this?** `var b strings.Builder` is required to get
> a ready-to-use, empty Builder (its zero value works). The name `b` is
> convention. Using `strings.Builder` means `"strings"` must be imported.

### Rune arithmetic

Because runes are numbers, you can do maths on them. In ASCII, every
lower-case letter sits a fixed distance after its upper-case one, and that
distance is `'a' - 'A'`:

```go
r := 'B'
r += 'a' - 'A'   // upper to lower: r is now 'b'
r -= 'a' - 'A'   // lower to upper: r is 'B' again
```

### Remember, then act later

Sometimes you can't decide what to write the moment you see something. Keep a
`bool` *flag* that remembers it, and act when the next relevant thing arrives:

```go
sawComma := false
for _, r := range "a,,b" {
	if r == ',' {
		sawComma = true   // note it, write nothing yet
		continue
	}
	// later: check sawComma here, then reset it to false
}
```

> **Why is it written like this?** The flag must be declared before the loop
> so it keeps its value between rounds. Its name is your choice.

### Cutting ASCII is safe

`s[:n]` counts bytes. For human text that's dangerous, but when a string
contains only ASCII, one byte is one character, so `s[:n]` is exactly the
first `n` characters.

## Your task

Your platform's "create service" form takes a display name, and the tooling
needs a DNS-safe name for the Kubernetes Service.

### 1. Slugify a display name

**Goal:** make `Slug(name)` return a lower-case ASCII name with single `-` between the kept parts.

- ASCII letters are kept, in lower case (`A`–`Z` become `a`–`z`).
- ASCII digits `0`–`9` are kept.
- Every other character is a separator: spaces, punctuation, `_`, `-`, emoji,
  and non-ASCII letters like `é`.
- Each run of one or more separators *between* kept characters becomes one `-`.
- Separators at the start or end produce nothing: the result never starts or
  ends with `-`.
- If nothing is kept, the result is `""`.

```go
Slug("Payments API")           // => "payments-api"
Slug("orders")                 // => "orders"
Slug("  Billing -- Worker!! ") // => "billing-worker"
Slug("My_Service.v2")          // => "my-service-v2"
Slug("eu-west-1 Gateway")      // => "eu-west-1-gateway"
Slug("🚀 Launch Service 🚀")    // => "launch-service"
Slug("Café Orders")            // => "caf-orders"
Slug("Zürich Edge")            // => "z-rich-edge"
Slug("***")                    // => ""
Slug("")                       // => ""
```

Yes, `"Café"` becomes `"caf"`: `é` is not ASCII, so it's a separator. Add
`"strings"` to the imports.

**Think it through:**
- For each rune: is it kept or a separator? Upper-case letters need turning
  into lower case first.
- Don't write `-` when you see a separator. Remember it, and decide when the
  next kept character arrives.
- How can `b.Len()` tell you whether this would be a leading `-`?

### 2. Make a valid DNS label

**Goal:** make `DNSLabel(name)` return `Slug(name)` shortened to a valid DNS label, or `"unnamed"`.

Apply these rules to the slug, in this order:

1. If it's longer than 63 characters, keep only the first 63.
2. If that leaves a `-` at the end, remove it (there's at most one; `Slug`
   never makes two in a row).
3. If the result is empty, return `"unnamed"`.

```go
DNSLabel("Payments API")                      // => "payments-api"
DNSLabel("Search 🔍 Indexer")                 // => "search-indexer"
DNSLabel("!!!")                               // => "unnamed"
DNSLabel("日本語")                             // => "unnamed"
DNSLabel("")                                  // => "unnamed"
DNSLabel(strings.Repeat("b", 63))             // => 63 "b"s (unchanged)
DNSLabel(strings.Repeat("a", 70))             // => 63 "a"s
DNSLabel(strings.Repeat("a", 62) + " backup") // => 62 "a"s (the cut left a "-")
DNSLabel(strings.Repeat("x", 60) + " Backup") // => 60 "x"s + "-ba"
```

**Think it through:**
- Call `Slug` first and keep its result in a variable you can change.
- The slug is pure ASCII: which kind of cut is safe now?
- The `strings` package has a function that removes a suffix only if it's
  there.

## Learn more

- [Kubernetes: DNS label names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names)
- [Package strings: Builder](https://pkg.go.dev/strings#Builder)
- [Package strings: TrimSuffix](https://pkg.go.dev/strings#TrimSuffix)
- [Go blog: Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
