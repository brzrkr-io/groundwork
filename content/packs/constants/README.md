# Constants

Some values should never change while a program runs: the default port, the
service name, the replica limit. Go lets you mark them as *constants*, and the
compiler then refuses any code that tries to change them.

**New here:** `const`, `const ( ... )` blocks, `int`, untyped constants.

## Learn

### Declaring a constant

A constant looks like a variable, but uses `const` and gets its value
straight away:

```go
const RetryLimit = 5
const Region = "eu-west-1"
```

`5` is a whole number. Go's type for whole numbers is `int`. `+` adds two
`int`s and `-` subtracts them. More on numbers in *Numbers*.

> **Why is it written like this?**
> - **Go requires** a value after `=` right away. A constant can't be filled
>   in later.
> - **Go allows** `const` outside functions (like here, at the top of the
>   file) and inside them. `:=` only works inside functions, but `const` and
>   `var` work in both places, because they start with a keyword.
> - **Convention:** shared settings go at the top of the file, above the
>   functions that use them.

### Grouping constants

Related constants usually go in one *const block*, one per line:

```go
// Settings for the billing service.
const (
	QueueName  = "invoices"
	MaxWorkers = 4
)
```

You use a constant exactly like a variable, by its name:

```go
QueueName + "-dlq" // => "invoices-dlq"
MaxWorkers - 1     // => 3
```

> **Why is it written like this?**
> - **Go requires** the `(` on the `const` line and the `)` on its own line
>   at the end. Inside, each line is `Name = value`, with no `const` in front.
> - **Convention:** gofmt lines up the `=` signs. You don't have to.
> - **Your choice:** whether to group constants, and their order.

Names that start with a capital letter, like `QueueName`, are *exported*:
code in other packages can use them. Lowercase names stay private to the
package. Both work the same inside your exercise.

### Untyped constants

`const MaxWorkers = 4` has no type written down. That makes it an *untyped
constant*: Go picks its type when it is used. You can write a type
(`const MaxWorkers int = 4`), but most Go code leaves it out. You'll see why
this helps in *Numbers*.

### Constants can't change

Assigning to a constant is a compile error:

```go
const RetryLimit = 5

func Retries() int {
	RetryLimit = 6 // does not compile
	return RetryLimit
}
```

```text
./exercise.go:4:2: cannot assign to RetryLimit (neither addressable nor a map index expression)
```

"Neither addressable" means "this isn't somewhere you can store a value". A
constant is a fixed value, not a box. When you need a different number,
**compute it and return it**, for example `RetryLimit + 1`. The constant
itself stays the same.

Adding `const` to that line doesn't help either: it would try to create a
second `RetryLimit`. Constants are only declared once.

## Your task

The checkout service's config helpers use a few constants. The starter does
**not compile**, so every test shows ✗ until you finish task 1.

### 1. Fix the compile error in SidecarPort

**Goal:** make `SidecarPort` return `DefaultPort` plus `n`, without changing
`DefaultPort`.

```go
SidecarPort(1)  // => 8081
SidecarPort(2)  // => 8082
SidecarPort(10) // => 8090
```

The error points at line 21, the line under `TODO(1)`. It tries to store a
new value in `DefaultPort`. The tests also check `DefaultPort` is still 8080.

**Think it through:**
- Do you need to store the new port anywhere, or only return it?
- Which two lines in `SidecarPort` could become one?

### 2. Build the in-cluster DNS name

**Goal:** make `DNSName` return
`<service>.<namespace>.svc.<cluster domain>`.

```go
DNSName("cluster.local")
// => "checkout.payments.svc.cluster.local"

DNSName("k8s.internal")
// => "checkout.payments.svc.k8s.internal"
```

The pieces: the `ServiceName` constant, a dot, the `Namespace` constant, the
fixed text `.svc.`, then the `clusterDomain` parameter. This is how
Kubernetes names services inside a cluster.

**Think it through:**
- The constants are used by name, like parameters. Do they need quotes?
- Which pieces are fixed text?

### 3. Count the spare replicas

**Goal:** make `SpareReplicas` return how many more replicas fit under
`MaxReplicas` (10) when `running` replicas are already up.

```go
SpareReplicas(3)  // => 7
SpareReplicas(10) // => 0
SpareReplicas(0)  // => 10
```

**Think it through:**
- If 3 of 10 are running, how do you get 7? Which operator is that?
- Which name holds the 10, and which holds the running count?

## Learn more

- [A Tour of Go: Constants](https://go.dev/tour/basics/15)
- [Go by Example: Constants](https://gobyexample.com/constants)
- [The Go Blog: Constants](https://go.dev/blog/constants)
- [Kubernetes: DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
