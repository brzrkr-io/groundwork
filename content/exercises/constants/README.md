# Constants

Some values should never change while a program runs: the default port, the
service name, the replica limit. Go lets you mark them as *constants*, and the
compiler then refuses any code that tries to change them.

## Learn

### Declaring a constant

A constant looks like a variable, but uses `const` and must get its value
straight away:

```go
const DefaultPort = 8080
const Region = "eu-west-1"
```

`8080` is a whole number. Go's type for whole numbers is `int`. You'll learn
more about numbers in the next exercise; for now, `+` adds two `int`s and `-`
subtracts them, as you'd expect.

### Grouping constants

Related constants usually go in one block, with one per line:

```go
// Settings for the checkout service.
const (
	ServiceName = "checkout"
	Namespace   = "payments"
	MaxReplicas = 10
)
```

You use a constant exactly like a variable:

```go
ServiceName + "-worker" // => "checkout-worker"
MaxReplicas - 4         // => 6
```

Names that start with a capital letter, like `ServiceName`, are *exported*:
code in other packages can use them. Lowercase names stay private to the
package. You'll see this again later; for now, both kinds work the same inside
your exercise.

### Untyped constants

`const MaxReplicas = 10` has no type written down. That makes it an *untyped
constant*: Go gives it a type only when it is used, to fit whatever it's used
with. This is why one numeric constant can work with several kinds of number.
You'll see that in *Numbers*.

You can also give a constant a type explicitly: `const MaxReplicas int = 10`.
Most Go code leaves it out.

### Constants can't change

Trying to assign to a constant is a compile error:

```go
const DefaultPort = 8080

func main() {
	DefaultPort = 9090 // does not compile
}
```

```text
./main.go:4:2: cannot assign to DefaultPort (neither addressable nor a map index expression)
```

"Neither addressable" is compiler language for "this isn't somewhere you can
store a value". A constant is a fixed value, not a box like a variable. If you
need a different number, compute it with an expression such as
`DefaultPort + 1`, or store it in a variable.

This is the point of constants: you can read `DefaultPort` anywhere in the
code and trust that it is 8080.

## Your task

The checkout service's config helpers use a few constants. The starter does
**not compile**, so every test shows ✗ until you finish task 1.

### 1. Fix the compile error in SidecarPort

Sidecar containers listen on the ports just above `DefaultPort`: sidecar 1
uses 8081, sidecar 2 uses 8082, and so on. `SidecarPort` tries to change
`DefaultPort` to get there, which Go doesn't allow. Read the compiler error,
then fix the function so it returns the sidecar's port without changing the
constant.

```go
SidecarPort(1)  // => 8081
SidecarPort(2)  // => 8082
SidecarPort(10) // => 8090
```

### 2. Build the in-cluster DNS name

Inside a Kubernetes cluster, a service is reachable at
`<service>.<namespace>.svc.<cluster domain>`. Change `DNSName` so it builds
that name from the `ServiceName` and `Namespace` constants and its
`clusterDomain` parameter:

```go
DNSName("cluster.local")
// => "checkout.payments.svc.cluster.local"

DNSName("k8s.internal")
// => "checkout.payments.svc.k8s.internal"
```

### 3. Count the spare replicas

The autoscaler may run at most `MaxReplicas` (10) replicas. Change
`SpareReplicas` so it returns how many more replicas could still be started
when `running` are already up:

```go
SpareReplicas(3)  // => 7
SpareReplicas(10) // => 0
SpareReplicas(0)  // => 10
```

## Learn more

- [A Tour of Go: Constants](https://go.dev/tour/basics/15)
- [Go by Example: Constants](https://gobyexample.com/constants)
- [The Go Blog: Constants](https://go.dev/blog/constants)
- [Kubernetes: DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
