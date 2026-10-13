# Stringers: a kubectl-style deployment report

`kubectl get deployments` turns a pile of structs into a few readable lines.
In this exercise you build the same kind of output for your own rollout
tool: small `String` methods that combine into one report.

## Learn

### Recap

Any type with a `String() string` method is a `fmt.Stringer`, and `fmt`
calls it for `%v`, `%s`, `Print`, `Println` and `Sprint`. Use a value
receiver so both values and pointers print nicely, and never format the
receiver itself with `%v` inside its own `String`, as in *Stringers*.

### Stringers inside stringers

When a struct's fields have their own `String` methods, the struct's
`String` can simply print the fields with `%v`. Each field formats itself:

```go
func (c Container) String() string {
	// c.Image and c.State are Stringers.
	return fmt.Sprintf("%s %v %v", c.Name, c.Image, c.State)
}
```

That is fine: `%v` on a *field* calls the field's `String`, not the
container's. Only `%v` on `c` itself would loop forever.

### Building text with strings.Builder

A `strings.Builder` collects text piece by piece, and `b.String()` returns
the result. `fmt.Fprintf` and `fmt.Fprintln` work like `Sprintf` and
`Println`, but they write into the builder you pass as the first argument
(with `&`) instead of returning a string:

```go
var b strings.Builder
fmt.Fprintf(&b, "hosts: %d\n", 2)
fmt.Fprintln(&b, "web-1") // Fprintln adds the "\n" for you
fmt.Fprintln(&b, "web-2")
b.String()
// => "hosts: 2\nweb-1\nweb-2\n"
```

`Fprintln` calls `String` just like `Println` does, so
`fmt.Fprintln(&b, deployment)` writes your formatted deployment and a
newline. (`Fprintf` accepts any `io.Writer`: an interface with a `Write`
method that `*strings.Builder` satisfies.)

## Your task

`Status` already prints itself as `Progressing`, `Available` or `Degraded`.
Add the rest of the output for your rollout tool.

### 1. Print replica counts

Add a `String` method to `Replicas` that returns the ready count, a slash and
the desired count:

```go
fmt.Sprint(Replicas{Ready: 3, Desired: 5})
// => "3/5"
```

### 2. Print a deployment

Add a `String` method to `Deployment` that returns the namespace, a slash,
the name, then the replicas and the status, separated by single spaces:

```go
d := Deployment{
	Name:      "api",
	Namespace: "prod",
	Replicas:  Replicas{Ready: 1, Desired: 3},
	Status:    Degraded,
}
fmt.Sprint(d)
// => "prod/api 1/3 Degraded"
```

### 3. Build the report

Finish `Report`. It returns a summary line followed by one line per
deployment, in the order given. Every line ends with `"\n"`:

```text
deployments: 3, degraded: 1
prod/api 3/3 Available
prod/web 1/2 Progressing
staging/worker 0/1 Degraded
```

- The summary is `deployments: N, degraded: M`, where `N` is the number of
  deployments and `M` is how many have status `Degraded`.
- Each deployment line is exactly what its `String` method returns.
- With no deployments, `Report` returns only the summary:
  `"deployments: 0, degraded: 0\n"`.

## Learn more

- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [strings.Builder](https://pkg.go.dev/strings#Builder)
- [fmt.Fprintf](https://pkg.go.dev/fmt#Fprintf)
- [Kubernetes: Deployment status](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#deployment-status)
