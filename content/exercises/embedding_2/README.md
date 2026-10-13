# Struct embedding: fix the Secret bugs

Embedding saves typing, but it has two classic traps: an outer field that hides
the embedded one, and an override that forgets to call the inner method. Both
bugs are in this exercise, in a model of Kubernetes Secrets.

## Learn

### Recap

As in *Struct embedding*, you embed a struct by
writing its type name with no field name. Its fields and methods are promoted,
so `s.Key()` is short for
`s.ObjectMeta.Key()`. In a literal you write the embedded struct out in full:
`ObjectMeta: ObjectMeta{Name: "db", Namespace: "prod"}`. An outer method with
the same name overrides the promoted one, and `s.ObjectMeta.Validate()` still
reaches the inner method.

### Name conflicts: the outer name wins

What happens if the outer struct has a field with the same name as a field in
the embedded struct?

```go
type Secret struct {
	ObjectMeta        // has a Name field
	Name       string // a second, separate Name
}

s := Secret{ObjectMeta: ObjectMeta{Namespace: "prod"}, Name: "db"}
s.Name            // "db": the outer field wins
s.ObjectMeta.Name // "": the embedded one was never set
s.Key()           // "prod/": Key only sees ObjectMeta
```

Go doesn't complain. The outer `Name` *shadows* the embedded one: `s.Name` now
means the outer field, and the embedded `Name` is only reachable as
`s.ObjectMeta.Name`. Methods promoted from `ObjectMeta` know nothing about the
outer field, so they see an empty name. The fix is to keep one `Name`, in
`ObjectMeta`.

## Your task

Your cluster tool stores Secrets (passwords, tokens, keys) as a `Secret` that
embeds `ObjectMeta`. Two bugs slipped in during review.

### 1. Remove the shadowing field

A teammate added a `Name` field to `Secret` without noticing that `ObjectMeta`
already has one. Now every secret has an empty key:

```go
s := NewSecret("prod", "db-password", map[string]string{"password": "hunter2"})
s.Key()
// => "prod/" (wrong)
```

Delete the extra `Name` field from `Secret`, and change `NewSecret` to put the
name inside `ObjectMeta`. Afterwards:

```go
s.Key()           // => "prod/db-password"
s.ObjectMeta.Name // => "db-password"
s.Data            // => map[password:hunter2]
```

### 2. Make Validate check the metadata too

`Secret` overrides `Validate`, but the override forgets to call
`ObjectMeta`'s `Validate`, so a secret with no name passes. Fix
`Secret.Validate` so it:

1. Runs `ObjectMeta`'s `Validate` first, and returns its error if there is one
   (`"name is required"` or `"namespace is required"`).
2. Otherwise returns an error `"secret has no data"` if `Data` is empty.
3. Otherwise returns `nil`.

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

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Go spec: Selectors (how the shallowest name wins)](https://go.dev/ref/spec#Selectors)
- [Kubernetes: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
