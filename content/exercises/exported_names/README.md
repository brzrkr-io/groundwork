# Exported and unexported names

Every package has a public face and private insides. Go decides which is
which by the first letter of the name. Getting this right is how you stop
other code from reading a password it should never see.

**New here:** exported and unexported names, constructors, `package x_test`.

## Learn

### Capital letter means exported

A name that starts with a capital letter is *exported*: other packages can
use it. A name that starts with a lower-case letter is *unexported*: only
code in the same package can use it.

```go
package dns

const MaxTTL = 86400 // exported: other packages write dns.MaxTTL
const minTTL = 60    // unexported: only package dns can use it

// Valid is exported: other packages can call dns.Valid(ttl).
func Valid(ttl int) bool {
	return ttl >= minTTL && ttl <= MaxTTL
}
```

> **Why is it written like this?**
> - **Go requires** it: the capital letter *is* the rule. There are no
>   `public` or `private` keywords.
> - It applies to everything a package declares: functions, types,
>   constants, variables, struct fields and methods.
> - **Convention:** a doc comment starts with the exact name it describes,
>   capital letter and all: `// Valid reports ...`.

From another package, an unexported name is off limits:

```text
name minTTL not exported by package dns
r.ttl undefined (cannot refer to unexported field ttl)
```

### Why keep things unexported?

Everything you export is a promise. Other teams start calling it, and then
you can't rename or remove it without breaking them. Keep helpers
unexported so you can change them freely, and export only what callers need.

Go also notices unexported things that nothing uses. Your editor greys them
out as *unused*: nobody inside the package calls them, and nobody outside
can.

### Unexported fields and constructors

A struct can mix exported and unexported fields:

```go
type Record struct {
	Host string // anyone can read and set this
	ttl  int    // only package dns can touch this
}
```

Code in another package can't write `dns.Record{ttl: 300}`, because it
can't see `ttl`. So the package offers a *constructor*: an exported function
that builds the value for you.

```go
func NewRecord(host string, ttl int) Record {
	return Record{Host: host, ttl: ttl}
}
```

Methods on `Record` live in package `dns`, so they can still read `r.ttl`.
The value stays usable but private.

> **Why is it written like this?**
> - **Go requires** nothing special: a constructor is an ordinary function.
> - **Convention:** it is named `New` plus the type name, and lives next to
>   the type.
> - **Author's choice:** which fields to hide. Hide anything callers must not
>   change or read.

### External test packages

A test file can declare `package dns_test` instead of `package dns`:

```go
package dns_test

import "example.com/infra/dns"

r := dns.NewRecord("api.internal", 300)
```

> **Why is it written like this?**
> - **Go requires** the `_test` suffix for this. It makes a separate package
>   in the same folder: the one exception to "one folder, one package".
> - It imports your package like any outside caller, so it sees only
>   exported names. It tests your package the way your users will use it.

## Your task

The `creds` package stores logins for the deploy bot. Right now it hides
what callers need and exposes the secret token. The tests use
`package creds_test`, so they can only see exported names.

```text
.
├── go.mod
└── creds/
    ├── creds.go           package creds (you edit this)
    └── creds_test.go      package creds_test
```

Until task 1 is done, the test file doesn't compile, so you'll see compiler
errors like `undefined: creds.NewCredential` instead of a test checklist.
That's expected.

### 1. Export what callers need

**Goal:** rename `newCredential` to `NewCredential` and `masked` to `Masked`.

Change only the first letter of each. Update their doc comments too, so each
still starts with the name. Their bodies stay the same.

`TestNewCredential` then checks the user and the header:

```go
c := creds.NewCredential("deploy-bot", "s3cr3t-value")
c.User     // => "deploy-bot"
c.Header() // => "Bearer s3cr3t-value"
```

**Think it through:**
- The editor marks both as unused. Why can't the tests reach them?
- Each name appears in two places: the doc comment and the `func` line.

### 2. Make the token private

**Goal:** rename the `Token` field of `Credential` to `token`, and fix every use of it in `creds.go`.

Right now any package can read the secret. `TestTokenIsPrivate` checks that
`User` is the only exported field of `Credential`. `Header` must keep
working.

**Think it through:**
- Renaming a field breaks every line that uses the old name. Look for `Token:` and `.Token`.
- `NewCredential` has a *parameter* called `token` too. Which side of `token: token` is the field, and which is the parameter?

### 3. Mask the token

**Goal:** finish the unexported helper `mask`, which `Masked` uses to hide the token in logs.

- A token of 4 characters or fewer (including an empty one) becomes `****`.
- A longer token keeps its first 2 characters, followed by `****`.

The number of stars is always 4, so the log doesn't reveal the token length.

```go
creds.NewCredential("deploy-bot", "s3cr3t-value").Masked() // => "deploy-bot:s3****"
creds.NewCredential("ci", "abcde").Masked()                // => "ci:ab****"
creds.NewCredential("ci", "abcd").Masked()                 // => "ci:****"
```

The test can't call `mask` directly, because it's unexported. It tests
`Masked` instead.

**Think it through:**
- Two cases means one `if`. What do you compare the length with?
- How do you take the first 2 characters of a string? (Slicing, as in *Slices*, works on strings too.)

## Learn more

- [Go spec: Exported identifiers](https://go.dev/ref/spec#Exported_identifiers)
- [Effective Go: Names](https://go.dev/doc/effective_go#names)
- [Go Tour: Exported names](https://go.dev/tour/basics/3)
