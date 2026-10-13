# Structs

A deployment has a name, an image, a replica count, environment variables and
resource limits. A *struct* keeps all of those together in one value, so you
can pass "a deployment" around instead of five loose variables.

## Learn

### Declaring a struct type

`type` gives a new type a name. `struct { ... }` lists its *fields*, each with
a name and a type:

```go
// Service is one running service.
type Service struct {
	Name string
	Port int
}
```

Fields of the same type can share a line: `Name, Image string`.

### Creating a value

A *struct literal* builds a value. Write the type, then `{}` with
`Field: value` pairs:

```go
svc := Service{Name: "api", Port: 8080}
```

Any field you leave out gets its zero value, as in *Zero values*. So
`Service{Name: "api"}` has `Port` 0, and `Service{}` has every field at its
zero value: `Name` is `""` and `Port` is `0`.

### Reading and changing fields

Use a dot to reach a field, both to read it and to assign to it:

```go
fmt.Println(svc.Name) // api
svc.Port = 9090
if svc.Port == 0 {
	svc.Port = 8080
}
```

### Structs inside structs

A field can itself be a struct. Chain the dots to reach inside:

```go
type Limits struct {
	CPUMillis int
	MemoryMB  int
}

type Job struct {
	Name   string
	Limits Limits
}

j := Job{Name: "backup", Limits: Limits{CPUMillis: 500, MemoryMB: 256}}
j.Limits.MemoryMB = 512
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

Inside a `[]Service{...}` literal you can leave out the type name before each
`{`.

### Passing structs to functions

A struct is a value, like an `int`. Passing it to a function hands over a
copy. To give back a changed struct, change the copy and `return` it:

```go
func WithPort(s Service, port int) Service {
	s.Port = port
	return s
}
```

### Printing a struct

The `%+v` verb prints a struct with its field names, which is handy when
debugging. `%v` prints only the values:

```go
fmt.Printf("%+v\n", svc) // {Name:api Port:9090}
fmt.Printf("%v\n", svc)  // {api 9090}
```

### Comparing structs

You can compare two structs with `==` when every field can be compared. A
struct with a map field cannot use `==`, so the tests in this exercise check
fields one at a time.

## Your task

You're writing the part of a deploy tool that builds and looks up
deployments. The `Resources` and `Deployment` types are already declared in
`exercise.go`. Read them first.

### 1. Build a deployment with defaults

`NewDeployment(name, image)` returns a `Deployment` with these fields:

| Field | Value |
|---|---|
| `Name` | the `name` argument |
| `Image` | the `image` argument |
| `Replicas` | `1` |
| `Env` | an empty map made with `map[string]string{}` (not `nil`) |
| `Limits` | `Resources{CPUMillis: 250, MemoryMB: 128}` |

```go
NewDeployment("api", "registry.local/api:1.4.0")
// => {Name:api Image:registry.local/api:1.4.0 Replicas:1 Env:map[]
//     Limits:{CPUMillis:250 MemoryMB:128}}
```

An empty map, unlike a `nil` map, is ready for `d.Env["LOG_LEVEL"] = "info"`.

### 2. Fill in missing settings

`ApplyDefaults(d)` returns `d` with two defaults filled in:

- if `Replicas` is `0`, set it to `1`
- if `Limits.MemoryMB` is `0`, set it to `128`

Every other field, including `Limits.CPUMillis`, stays exactly as it was.
Values that are already set are kept.

```go
ApplyDefaults(Deployment{Name: "worker"})
// => {Name:worker Replicas:1 Limits:{CPUMillis:0 MemoryMB:128} ...}

ApplyDefaults(Deployment{Name: "db", Replicas: 3, Limits: Resources{CPUMillis: 1000, MemoryMB: 2048}})
// => {Name:db Replicas:3 Limits:{CPUMillis:1000 MemoryMB:2048} ...}
```

### 3. Find a deployment by name

`FindDeployment(ds, name)` looks through `ds` for the deployment whose `Name`
equals `name`. It returns that deployment and `true`. When no deployment has
that name, it returns the zero `Deployment` (`Deployment{}`) and `false`.

```go
ds := []Deployment{
	{Name: "api", Image: "api:1.4.0", Replicas: 3},
	{Name: "web", Image: "web:2.0.1", Replicas: 2},
}
FindDeployment(ds, "web")   // => {Name:web Image:web:2.0.1 Replicas:2 ...}, true
FindDeployment(ds, "cache") // => {}, false
```

## Learn more

- [A Tour of Go: Structs](https://go.dev/tour/moretypes/2)
- [Go by Example: Structs](https://gobyexample.com/structs)
- [Go spec: Struct types](https://go.dev/ref/spec#Struct_types)
- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
