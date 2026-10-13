# Struct embedding

Every Kubernetes object, from a Pod to a Service, carries the same metadata: a
name, a namespace and labels. Go lets you write that shared part once and
*embed* it in every object type.

## Learn

### Embedding a struct

To embed a struct, list its type name inside another struct, with no field
name in front:

```go
type Metadata struct {
	Name      string
	Namespace string
}

type Pod struct {
	Metadata        // embedded: no field name, just the type
	Image    string
}
```

A `Pod` now contains a whole `Metadata` value. The embedded field is named
after its type, so you can reach it as `p.Metadata`.

### Promoted fields

The fields of the embedded struct are *promoted*: you can use them as if they
belonged to `Pod` itself.

```go
var p Pod
p.Name = "api"            // short for p.Metadata.Name
fmt.Println(p.Metadata.Name) // api
```

Both spellings reach the same field. The short one is what you will use most
of the time.

### Writing the embedded struct in a literal

Promotion works for reading and assigning, but **not** in a struct literal.
In a literal, you fill in the embedded struct by its type name, like any other
field:

```go
p := Pod{
	Metadata: Metadata{Name: "api", Namespace: "default"},
	Image:    "nginx:1.27",
}
// Pod{Name: "api"} does not compile.
```

### Promoted methods

Methods are promoted too. A method declared on `Metadata` can be called on a
`Pod`:

```go
func (m Metadata) Key() string {
	return m.Namespace + "/" + m.Name
}

p.Key() // "default/api", short for p.Metadata.Key()
```

When Go runs `p.Key()`, it really calls `p.Metadata.Key()`. Inside `Key`, `m`
is the `Metadata`, not the `Pod`, so the method cannot see `p.Image`.

### Overriding a promoted method

If the outer type declares a method with the same name, the outer one wins.
You can still call the inner one by naming the embedded field:

```go
func (p Pod) Key() string {
	return "pod:" + p.Metadata.Key() // calls Metadata's Key
}
```

Writing `p.Key()` inside this method would call the `Pod` version again, which
calls itself forever. Always go through `p.Metadata`.

### Embedding is not inheritance

If you have used other languages, embedding may look like inheritance. It
isn't. A `Pod` *has* a `Metadata`; it *is not* one.

| You might expect | What Go does |
|---|---|
| A `Pod` can be passed where a `Metadata` is wanted | No. Pass `p.Metadata` instead. |
| `Metadata`'s methods see the `Pod` | No. They only see the `Metadata` part. |
| Fields and methods are copied into `Pod` | They are reachable by a shorter name. |

## Your task

You are writing a tiny model of Kubernetes objects for a deploy tool. `Pod`
and `Service` both embed `Metadata`.

### 1. Build a pod

Finish `NewPod` so it returns a `Pod` with:

- `Name` and `Namespace` set from the arguments,
- `Labels` holding exactly one label, `"app"`, set to the pod's name,
- `Image` set from the argument.

```go
p := NewPod("api", "default", "nginx:1.27")
p.Name          // => "api"
p.Namespace     // => "default"
p.Labels["app"] // => "api"
p.Image         // => "nginx:1.27"
```

The tests read the promoted fields (`p.Name`), so the values must live inside
the embedded `Metadata`.

### 2. Describe the metadata

Finish the `Describe` method on `Metadata`. It returns the namespace, a slash
and the name:

```go
m := Metadata{Name: "api", Namespace: "default"}
m.Describe()
// => "default/api"
```

`Service` does not declare `Describe`, so it gets this method through
promotion:

```go
s := Service{Metadata: Metadata{Name: "web", Namespace: "prod"}, Port: 80}
s.Describe()
// => "prod/web"
```

### 3. Override Describe for pods

Finish the `Describe` method on `Pod`. It returns `pod `, then the metadata
description from task 2, then the image in parentheses:

```go
p := Pod{
	Metadata: Metadata{Name: "api", Namespace: "default"},
	Image:    "nginx:1.27",
}
p.Describe()
// => "pod default/api (nginx:1.27)"
```

Reuse `Metadata`'s `Describe` instead of building `default/api` again.
If you use `fmt.Sprintf`, add `import "fmt"` below the `package` line.

## Learn more

- [Effective Go: Embedding](https://go.dev/doc/effective_go#embedding)
- [Go by Example: Struct Embedding](https://gobyexample.com/struct-embedding)
- [Go spec: Struct types (embedded fields and promotion)](https://go.dev/ref/spec#Struct_types)
- [Kubernetes: Object names and namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/)
