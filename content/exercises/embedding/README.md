# Struct embedding

Every Kubernetes object, from a Pod to a Service, carries the same metadata: a
name, a namespace and labels. Go lets you write that shared part once and
*embed* it in every object type.

**New here:** embedded fields, promoted fields and methods, overriding a promoted method.

## Learn

### Embedding a struct

To embed a struct, write its type name inside another struct, with **no field
name** in front:

```go
type Owner struct {
	Team  string
	Email string
}

type Job struct {
	Owner         // embedded: just the type, no field name
	Schedule string
}
```

A `Job` now contains a whole `Owner`. The embedded field is named after its
type, so you reach it as `j.Owner`.

> **Why is it written like this?**
> - **Go requires** a type with no name in front. `Owner Owner` would be a normal field, not an embedded one.
> - **Convention:** embedded fields go first, on their own line.
> - **Author's choice:** which structs you embed.

### Promoted fields

The embedded struct's fields are *promoted*: you can use them as if they
belonged to `Job`.

```go
var j Job
j.Team = "sre"          // short for j.Owner.Team
fmt.Println(j.Owner.Team) // sre
```

Both spellings reach the **same** field.

### Embedded structs in a literal

Promotion works for reading and assigning, but **not** in a struct literal.
There you fill in the embedded struct by its type name, like any field:

```go
j := Job{
	Owner:    Owner{Team: "sre", Email: "sre@example.com"},
	Schedule: "@daily",
}
// Job{Team: "sre"} does not compile.
```

> **Why is it written like this?**
> - **Go requires** `Owner: Owner{...}` in a literal. The left `Owner` is the field name, the right `Owner{...}` is the value.

### Promoted methods

Methods are promoted too. A method on `Owner` can be called on a `Job`:

```go
func (o Owner) Contact() string {
	return o.Team + " <" + o.Email + ">"
}

j.Contact() // "sre <sre@example.com>", short for j.Owner.Contact()
```

Inside `Contact`, `o` is only the `Owner`. It cannot see `j.Schedule`.

### Overriding a promoted method

If the outer type declares a method with the **same name**, the outer one
wins. You still reach the inner one through the embedded field:

```go
func (j Job) Contact() string {
	return "job owner: " + j.Owner.Contact() // calls Owner's Contact
}
```

Writing `j.Contact()` inside this method would call the `Job` version again,
which calls itself forever. Go through `j.Owner`.

### Embedding is not inheritance

A `Job` *has* an `Owner`; it *is not* one.

| You might expect | What Go does |
|---|---|
| A `Job` can be passed where an `Owner` is wanted | No. Pass `j.Owner`. |
| `Owner`'s methods see the `Job` | No. They see only the `Owner` part. |
| Fields and methods are copied into `Job` | They are just reachable by a shorter name. |

## Your task

You are modelling Kubernetes objects for a deploy tool. `Pod` and `Service`
both embed `Metadata` (fields `Name`, `Namespace`, `Labels`).

### 1. Build a pod

**Goal:** finish `NewPod` so the returned `Pod` has its name, namespace, labels and image filled in.

- `Name` and `Namespace` come from the parameters `name` and `namespace`.
- `Labels` holds exactly **one** label: key `"app"`, value = the pod's name.
- `Image` comes from the parameter `image` (already done).

```go
p := NewPod("api", "default", "nginx:1.27")
p.Name          // => "api"
p.Namespace     // => "default"
p.Labels["app"] // => "api"
p.Image         // => "nginx:1.27"
```

**Think it through:**
- `Name`, `Namespace` and `Labels` live in `Metadata`. How do you fill an embedded struct inside a literal?
- `Labels` is a map. How do you write a map literal with one key?
- `"app"` is fixed text; the value comes from a parameter.

### 2. Describe the metadata

**Goal:** finish `Metadata`'s `Describe` so it returns the namespace, a slash, then the name.

```go
Metadata{Name: "api", Namespace: "default"}.Describe()
// => "default/api"
```

`Service` has no `Describe` of its own, so it gets this one by promotion:

```go
s := Service{Metadata: Metadata{Name: "web", Namespace: "prod"}, Port: 80}
s.Describe()
// => "prod/web"
```

**Think it through:**
- Which receiver name does the method use? Read its fields through that.
- Namespace comes first. The `/` is fixed text.

### 3. Override Describe for pods

**Goal:** finish `Pod`'s `Describe` so it returns `pod `, the metadata description, then the image in parentheses.

```go
p := Pod{
	Metadata: Metadata{Name: "api", Namespace: "default"},
	Image:    "nginx:1.27",
}
p.Describe()
// => "pod default/api (nginx:1.27)"
```

**Think it through:**
- Reuse `Metadata`'s `Describe` instead of building `default/api` again. How do you call the inner one, not this one?
- Fixed text: `pod ` (with a space), ` (` and `)`.
- If you use `fmt.Sprintf`, add `import "fmt"` below the `package` line.

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Go by Example: Struct Embedding](https://gobyexample.com/struct-embedding)
- [Go spec: Struct types (embedded fields and promotion)](https://go.dev/ref/spec#Struct_types)
- [Kubernetes: Object names and namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/)
