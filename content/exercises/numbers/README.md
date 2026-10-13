# Numbers and arithmetic

Capacity planning is arithmetic: how much memory do the replicas need, how
many replicas handle the traffic, how busy is the CPU. Here you learn Go's
two everyday number types and the traps in dividing them.

## Learn

### Two kinds of number

| Type | Holds | Examples |
|---|---|---|
| `int` | whole numbers | `3`, `8080`, `-1` |
| `float64` | numbers with a fractional part | `0.5`, `99.9`, `2.0` |

Use `int` for things you count: replicas, ports, requests. Use `float64` for
measurements and ratios: CPU usage, percentages, prices.

### Arithmetic operators

| Operator | Meaning | `int` example | `float64` example |
|---|---|---|---|
| `+` | add | `7 + 2` is `9` | `7.0 + 2.0` is `9.0` |
| `-` | subtract | `7 - 2` is `5` | `7.0 - 2.0` is `5.0` |
| `*` | multiply | `7 * 2` is `14` | `7.0 * 2.0` is `14.0` |
| `/` | divide | `7 / 2` is `3` (!) | `7.0 / 2.0` is `3.5` |
| `%` | remainder | `7 % 2` is `1` | not allowed |

`*` and `/` happen before `+` and `-`, as in school maths. Use parentheses
to change the order: `(2 + 3) * 4` is 20.

### Integer division throws the fraction away

When both sides are `int`, `/` gives an `int`, and the fractional part is
dropped (*truncated*). It does not round:

```go
7 / 2   // => 3, not 3.5
999 / 1000 // => 0
```

`%` gives the remainder that was thrown away: `7 % 2` is 1, because 7 is
3 twos with 1 left over.

### Rounding up: the ceiling trick

Say each replica handles 300 requests per second and you get 1000. Then
`1000 / 300` is 3, but 3 replicas only handle 900. You need 4. You need to
round *up*, which is called the *ceiling*.

To round `a / b` up with whole numbers, add `b - 1` before dividing:

```go
(a + b - 1) / b
(1000 + 300 - 1) / 300 // => 1299 / 300 => 4
(900 + 300 - 1) / 300  // => 1199 / 300 => 3
```

If `a` divides exactly, the extra `b - 1` isn't enough to reach the next
number. If there is any remainder at all, it pushes the result up by one.

### int and float64 don't mix

Go never mixes the two types for you. If `a` is an `int` and `b` is a
`float64`, then `a + b` does not compile. You'll learn to convert between them
in *Type conversion*. In this exercise each function uses only one type.

## Your task

Write the capacity helpers for your team's planning dashboard.

### 1. Total the memory

Change `TotalMemoryMiB` so it returns the memory that `replicas` replicas need
when each one needs `perReplicaMiB` mebibytes (MiB):

```go
TotalMemoryMiB(3, 256) // => 768
TotalMemoryMiB(5, 512) // => 2560
TotalMemoryMiB(0, 512) // => 0
```

### 2. Round the replica count up

`ReplicasNeeded` returns how many replicas are needed to serve
`requestsPerSec` when each replica can handle `perReplica` requests per
second. The starter divides, which rounds down and leaves you short. Change it
to round up:

```go
ReplicasNeeded(1000, 300) // => 4
ReplicasNeeded(900, 300)  // => 3
ReplicasNeeded(1, 300)    // => 1
ReplicasNeeded(0, 300)    // => 0
```

`perReplica` is never 0.

### 3. Work out the CPU percentage

Kubernetes measures CPU in *millicores*: 1000 millicores is one full CPU.
Change `CPUPercent` so it returns how much of its CPU limit a container uses,
as a percentage: used divided by the limit, times 100.

```go
CPUPercent(250, 1000)  // => 25
CPUPercent(125, 1000)  // => 12.5
CPUPercent(1500, 1000) // => 150
```

A result above 100 means the container wants more CPU than its limit and is
being slowed down. `limitMillicores` is never 0.

## Learn more

- [A Tour of Go: Basic types](https://go.dev/tour/basics/11)
- [Go spec: Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators)
- [Kubernetes: CPU resource units](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-cpu)
