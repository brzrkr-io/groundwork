# Structs

A deployment has a name, an image, a replica count, environment variables and
resource limits. A *struct* keeps all of those together in one value, so you
pass around "a deployment" instead of five loose variables.

**New here:** `type ... struct`, struct literals, field access with `.`,
nested structs, slices of structs.

## Learn

### Declaring a struct type

`type` gives a new type a name. `struct { ... }` lists its *fields*, each a
name followed by a type:

```go
// Service is one running service.
type Service struct {
	Name string
	Port int
}
```

> **Why is it written like this?** Go requires `type`, the name, `struct`
> and the braces, and in each field the name comes *before* the type.
> Capitalised names (`Service`, `Name`) are visible to other packages: Go
> requires that rule. Lining up the types in a column is gofmt's job, not
> yours. The names themselves are the author's choice.

You don't declare any types in this exercise: `Resources` and `Deployment`
are already in `exercise.go`. Read them before you start.

### Creating a value

A *struct literal* builds a value: the type name, then `{}` holding
`Field: value` pairs separated by commas.

```go
svc := Service{Name: "api", Port: 8080}
```

Any field you leave out gets its zero value, as in *Zero values: default the config*.
`Service{Name: "api"}` has `Port` 0. `Service{}` has every field at zero:
`Name` is `""`, `Port` is `0`.

> **Why is it written like this?** The `Field: value` form and the commas
> are required. Field order inside the braces is your choice. When you
> spread a literal over several lines, Go requires a comma after the *last*
> field too, so adding a line later never breaks the one above.

### Reading and changing fields

A dot reaches a field, both to read it and to assign to it:

```go
fmt.Println(svc.Name) // api
svc.Port = 9090
if svc.Port == 0 {
	svc.Port = 8080
}
```

### Structs inside structs

A field can itself be a struct. Its value is then another struct literal,
and you chain dots to reach inside:

```go
type Disk struct {
	SizeGB int
	Type   string
}

type Volume struct {
	Name string
	Disk Disk
}

v := Volume{Name: "logs", Disk: Disk{SizeGB: 20, Type: "ssd"}}
v.Disk.SizeGB = 50
```

### Slices of structs

A slice can hold structs, and `range` walks through them as usual:

```go
services := []Service{
	{Name: "api", Port: 8080},
	{Name: "web", Port: 80},
}
for _, s := range services {
	fmt.Println(s.Name, s.Port)
}
```

Inside a `[]Service{...}` literal you may leave out the type name before each
`{`. That's allowed by Go, not required.

### Passing structs to functions

A struct is a value, like an `int`. Passing it to a function hands over a
*copy*. To give back a changed struct, change the copy and `return` it:

```go
func WithPort(s Service, port int) Service {
	s.Port = port
	return s
}
```

### Printing and comparing

`%+v` prints a struct with its field names; `%v` prints only the values. The
examples below use the `%+v` style.

```go
fmt.Printf("%+v\n", svc) // {Name:api Port:9090}
fmt.Printf("%v\n", svc)  // {api 9090}
```

You can compare two structs with `==` only when every field can be compared.
A map field can't, so the tests here check fields one at a time.

## Your task

You're writing the part of a deploy tool that builds and looks up
deployments.

### 1. Build a deployment with defaults

**Goal:** make `NewDeployment(name, image)` return a `Deployment` with every
field filled in.

| Field | Value |
|---|---|
| `Name` | the `name` parameter |
| `Image` | the `image` parameter |
| `Replicas` | `1` |
| `Env` | an empty map of `string` to `string` (not `nil`) |
| `Limits` | a `Resources` with `CPUMillis` 250 and `MemoryMB` 128 |

```go
NewDeployment("api", "registry.local/api:1.4.0")
// => {Name:api Image:registry.local/api:1.4.0 Replicas:1 Env:map[]
//     Limits:{CPUMillis:250 MemoryMB:128}}
```

An empty map, unlike a `nil` map, is ready for `d.Env["LOG_LEVEL"] = "info"`.

**Think it through:**
- Which two values come from the parameters, and which three are fixed?
- `Limits` has type `Resources`. What does a value of a struct type look
  like (see *Structs inside structs*)?
- How did you write an empty map literal in *Maps*?

### 2. Fill in missing settings

**Goal:** make `ApplyDefaults(d)` return `d` with two missing values filled
in.

- If `Replicas` is `0`, set it to `1`.
- If `Limits.MemoryMB` is `0`, set it to `128`.

Every other field, including `Limits.CPUMillis`, stays exactly as it was.
Values that are already set are kept.

```go
ApplyDefaults(Deployment{Name: "worker"})
// => {Name:worker Replicas:1 Limits:{CPUMillis:0 MemoryMB:128} ...}

ApplyDefaults(Deployment{Name: "db", Replicas: 3, Limits: Resources{CPUMillis: 1000, MemoryMB: 2048}})
// => {Name:db Replicas:3 Limits:{CPUMillis:1000 MemoryMB:2048} ...}
```

**Think it through:**
- `d` is already your own copy. You can change its fields and keep the
  existing `return d`.
- Memory lives one level down, inside `Limits`. How many dots does that
  take?

### 3. Find a deployment by name

**Goal:** make `FindDeployment(ds, name)` return the deployment whose `Name`
equals `name`, plus `true`.

When no deployment has that name (or `ds` is empty), return the zero
`Deployment` (`Deployment{}`) and `false`. The starter's last line already
does that, so keep it.

```go
ds := []Deployment{
	{Name: "api", Image: "api:1.4.0", Replicas: 3},
	{Name: "web", Image: "web:2.0.1", Replicas: 2},
}
FindDeployment(ds, "web")   // => {Name:web Image:web:2.0.1 Replicas:2 ...}, true
FindDeployment(ds, "cache") // => {}, false
```

**Think it through:**
- You need to look at each deployment in `ds` in turn. Which loop does
  that?
- Compare which field with which parameter?
- Return as soon as you find a match; the "not found" line runs only if the
  loop finishes.

## Learn more

- [A Tour of Go: Structs](https://go.dev/tour/moretypes/2)
- [Go by Example: Structs](https://gobyexample.com/structs)
- [Go spec: Struct types](https://go.dev/ref/spec#Struct_types)
- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
