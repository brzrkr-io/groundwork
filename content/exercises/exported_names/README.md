# Exported and unexported names

Every package has a public face and private insides. Go decides which is
which with one simple rule: the first letter of the name. Getting this right
is how you stop other code from reading a password it should never see.

## Learn

### Capital letter means exported

A name that starts with a capital letter is *exported*: other packages can
use it. A name that starts with a lower-case letter is *unexported*: only
code in the same package can use it.

```go
package creds

const MaxTokenLength = 256 // exported
const minTokenLength = 8   // unexported

// Valid is exported: other packages can call creds.Valid(token).
func Valid(token string) bool {
	return len(token) >= minTokenLength && len(token) <= MaxTokenLength
}

// tooShort is unexported: a helper for code inside package creds only.
func tooShort(token string) bool {
	return len(token) < minTokenLength
}
```

The rule applies to everything a package declares: functions, types,
constants, variables, struct fields and methods. There are no `public` or
`private` keywords.

From another package, an unexported name is off limits, and the compiler
tells you so:

```text
name mask not exported by package creds
c.masked undefined (cannot refer to unexported method masked)
```

The tests in this exercise call names that don't exist yet, so you'll see a
related message: `undefined: creds.NewCredential`.

### Why keep things unexported?

Everything you export is a promise. Other teams start calling it, and then
you can't rename or remove it without breaking them. Keep helpers like
`mask` unexported so you can change them freely, and export only what callers
need.

### Unexported fields and constructors

A struct can mix exported and unexported fields:

```go
type Credential struct {
	User  string // anyone can read and set this
	token string // only package creds can touch this
}
```

Code in another package can't write `creds.Credential{token: "x"}`, because
it can't see `token`. So the package offers a *constructor*: an exported
function, named `New` plus the type name by convention, that builds the value
for you:

```go
func NewCredential(user, token string) Credential {
	return Credential{User: user, token: token}
}
```

Methods on `Credential` live in package `creds`, so they can still read
`c.token`. The token stays usable but private.

### External test packages

A test file can declare `package creds_test` instead of `package creds`. The
`_test` suffix makes it a separate package that imports `creds` like any
other caller:

```go
package creds_test

import "gopherlings.local/exported_names/creds"

c := creds.NewCredential("deploy-bot", "s3cr3t")
```

It sees only exported names, so it tests your package exactly the way your
users will. The tests in this exercise work this way.

## Your task

The `creds` package stores logins for the deploy bot. Right now it hides
what callers need and exposes the secret token. The tests in
`creds/creds_test.go` use `package creds_test`.

```text
.
├── go.mod
└── creds/
    ├── creds.go           package creds (you edit this)
    └── creds_test.go      package creds_test
```

Until task 1 is done, the test file doesn't compile, so you'll see compiler
errors instead of a test checklist. That's expected.

### 1. Export what callers need

The tests call `creds.NewCredential` and `c.Masked()`. Rename:

| Now | Rename to |
|---|---|
| `newCredential` | `NewCredential` |
| `masked` | `Masked` |

Update the doc comments too: by convention they start with the name.

Your editor marks `newCredential` and `masked` as unused. Nothing inside
`creds` calls them, and code outside can't, so as far as Go can tell they are
dead code. The warning goes away once they're exported.
`TestNewCredential` then checks the user and the header:

```go
c := creds.NewCredential("deploy-bot", "s3cr3t-value")
c.User     // => "deploy-bot"
c.Header() // => "Bearer s3cr3t-value"
```

### 2. Make the token private

`Credential` has an exported `Token` field, so any package can read the
secret. Rename it to `token`, and fix every place in `creds.go` that uses it.
`TestTokenIsPrivate` checks that `User` is the only exported field of
`Credential`.

### 3. Mask the token

Finish the unexported helper `mask`, which `Masked` uses:

- A token of 4 characters or fewer becomes `****`.
- A longer token keeps its first 2 characters, followed by `****`.

The number of stars is always 4, so the log doesn't reveal the token length.

```go
creds.NewCredential("deploy-bot", "s3cr3t-value").Masked() // => "deploy-bot:s3****"
creds.NewCredential("ci", "abcdef").Masked()               // => "ci:ab****"
creds.NewCredential("ci", "abc").Masked()                  // => "ci:****"
```

The test can't call `mask` directly, because it's unexported. It tests
`Masked` instead.

## Learn more

- [Go spec: Exported identifiers](https://go.dev/ref/spec#Exported_identifiers)
- [Effective Go: Names](https://go.dev/doc/effective_go#names)
- [Go Tour: Exported names](https://go.dev/tour/basics/3)
