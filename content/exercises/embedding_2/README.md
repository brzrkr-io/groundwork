# Struct embedding: fix the Secret bugs

Embedding has two classic traps: an outer field that hides the embedded one,
and an override that forgets to call the inner method. Both bugs are in this
model of Kubernetes Secrets.

**New here:** shadowing (an outer name hiding an embedded one).

## Learn

### Recap

As in *Struct embedding*: embed a struct by writing its type name with no
field name. Its fields and methods are *promoted*, so `x.Field` is short for
`x.Inner.Field`.

- In a literal, write the embedded struct out in full: `Inner: Inner{...}`.
- An outer method with the same name *overrides* the promoted one.
- `x.Inner.Method()` still reaches the inner method.

### Name conflicts: the outer name wins

What if the outer struct has a field with the **same name** as one in the
embedded struct?

```go
type Labels struct {
	Team string
}

func (l Labels) Owner() string { return "team " + l.Team }

type Job struct {
	Labels         // has a Team field
	Team   string  // a second, separate Team
}

j := Job{Team: "sre"}
j.Team        // "sre": the outer field wins
j.Labels.Team // "": the embedded one was never set
j.Owner()     // "team ": Owner only sees Labels
```

Go doesn't complain. The outer `Team` *shadows* the embedded one. Promoted
methods know nothing about the outer field, so they see an empty value.

The fix: keep **one** field, inside the embedded struct.

> **Why is it written like this?**
> - **Go requires** the shallowest name to win: `j.Team` picks the field closest to `Job`. This is a language rule, not a bug Go will report.
> - **Convention:** never repeat an embedded field's name in the outer struct.

## Your task

Your cluster tool stores Secrets (passwords, tokens, keys) as a `Secret` that
embeds `ObjectMeta` (fields `Name`, `Namespace`). Two bugs slipped in.

### 1. Remove the shadowing field

**Goal:** delete `Secret`'s extra `Name` field and make `NewSecret` store the name inside `ObjectMeta`.

Right now every secret has an empty key:

```go
s := NewSecret("prod", "db-password", map[string]string{"password": "hunter2"})
s.Key()
// => "prod/" (wrong)
```

After the fix:

```go
s.Key()           // => "prod/db-password"
s.ObjectMeta.Name // => "db-password"
s.Data            // => map[password:hunter2]
```

**Think it through:**
- Look at the `TODO(1)` in the `Secret` type: which line duplicates a field `ObjectMeta` already has?
- Once that line is gone, the `Name:` line in `NewSecret` won't compile. Where should the `name` parameter go instead?

### 2. Make Validate check the metadata too

**Goal:** fix `Secret.Validate` so it checks `ObjectMeta`'s rules first, then the data.

In this order:

1. Run `ObjectMeta`'s `Validate`. If it returns an error (`"name is required"`
   or `"namespace is required"`), return that error.
2. Otherwise, if `Data` is empty, return an error `"secret has no data"`.
3. Otherwise return `nil`.

```go
Secret{ObjectMeta: ObjectMeta{Namespace: "prod"}, Data: map[string]string{"token": "abc"}}.Validate()
// => error "name is required"

Secret{ObjectMeta: ObjectMeta{Name: "api-token"}}.Validate()
// => error "namespace is required" (metadata is checked before data)

Secret{ObjectMeta: ObjectMeta{Name: "api-token", Namespace: "prod"}}.Validate()
// => error "secret has no data"

Secret{ObjectMeta: ObjectMeta{Name: "api-token", Namespace: "prod"}, Data: map[string]string{"token": "abc"}}.Validate()
// => nil
```

**Think it through:**
- `s.Validate()` inside this method would call itself. How do you reach the *inner* `Validate`?
- Don't write new error messages for the metadata: pass on the one you get back.
- The data check below already works; your code goes above it.

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Go spec: Selectors (how the shallowest name wins)](https://go.dev/ref/spec#Selectors)
- [Kubernetes: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
