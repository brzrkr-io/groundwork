# Numbers and arithmetic

Capacity planning is arithmetic: how much memory do the replicas need, how
many replicas handle the traffic, how busy is the CPU. Here you learn Go's
two everyday number types and the traps in dividing them.

**New here:** `int`, `float64`, `+ - * / %`, integer division, rounding up.

## Learn

### Two kinds of number

| Type | Holds | Examples |
|---|---|---|
| `int` | whole numbers | `3`, `8080`, `-1` |
| `float64` | numbers with a fractional part | `0.5`, `99.9`, `2.0` |

Use `int` for things you count: replicas, ports, requests. Use `float64` for
measurements and ratios: CPU usage, percentages, prices.

Numbers are written **without quotes**. `256` is a number; `"256"` is text.

### Reading a number function's `func` line

```text
func   Average   (total, count float64)   float64   {
keyword  name     parameters and type     result   body
```

Both parameters are `float64`, and so is the result. In the body, you use
`total` and `count` by name to do the maths, and `return` the answer.

> **Why is it written like this?**
> - **Go requires** the type after the names (`total, count float64`). Go
>   puts types after names so a line reads left to right: "total and count,
>   both float64".
> - **Go requires** the result type after the `)`. The `return` value must
>   be that type: return an `int` from a `float64` function and it won't
>   compile.
> - **Your choice:** the names. Units in names (`MiB`, `Millicores`) are a
>   common habit in ops code, so nobody mixes up MiB and GiB.

### Arithmetic operators

| Operator | Meaning | `int` example | `float64` example |
|---|---|---|---|
| `+` | add | `7 + 2` is `9` | `7.0 + 2.0` is `9.0` |
| `-` | subtract | `7 - 2` is `5` | `7.0 - 2.0` is `5.0` |
| `*` | multiply | `7 * 2` is `14` | `7.0 * 2.0` is `14.0` |
| `/` | divide | `7 / 2` is `3` (!) | `7.0 / 2.0` is `3.5` |
| `%` | remainder | `7 % 2` is `1` | not allowed |

`*` and `/` happen before `+` and `-`, as in school maths. Parentheses change
the order: `(2 + 3) * 4` is 20.

> **Why is it written like this?** **Convention:** gofmt sometimes removes
> spaces around `*` and `/` to show they happen first, as in `a + b*c`. The
> spaces don't change the maths; only parentheses do.

### Integer division throws the fraction away

When both sides are `int`, `/` gives an `int`, and the fractional part is
dropped (*truncated*). It does not round:

```go
7 / 2      // => 3, not 3.5
999 / 1000 // => 0
```

`%` gives the remainder that was thrown away: `7 % 2` is 1, because 7 is
3 twos with 1 left over.

### Rounding up: the ceiling trick

Say each worker handles 40 jobs and you have 100 jobs. `100 / 40` is 2, but 2
workers only handle 80. You need 3. You need to round *up*, which is called
the *ceiling*.

To round `a / b` up with whole numbers, add `b - 1` to `a` before dividing:

```go
(100 + 40 - 1) / 40 // => 139 / 40 => 3
(80 + 40 - 1) / 40  // => 119 / 40 => 2
```

If `a` divides exactly, the extra `b - 1` isn't enough to reach the next
number. If there is any remainder, it pushes the result up by one. The
parentheses matter: without them, `/` would run first.

### int and float64 don't mix

Go never mixes the two types for you. If `a` is an `int` and `b` is a
`float64`, then `a + b` does not compile. You'll learn to convert in *Type
conversion*. In this exercise each function uses only one type.

## Your task

Write the capacity helpers for your team's planning dashboard.

### 1. Total the memory

**Goal:** make `TotalMemoryMiB` return the memory `replicas` replicas need
when each needs `perReplicaMiB` mebibytes (MiB).

```go
TotalMemoryMiB(3, 256) // => 768
TotalMemoryMiB(5, 512) // => 2560
TotalMemoryMiB(0, 512) // => 0
```

**Think it through:**
- 3 replicas of 256 each: which operator gets 768?
- Use the two parameter names, not the numbers from the example.

### 2. Round the replica count up

**Goal:** make `ReplicasNeeded` return how many replicas serve
`requestsPerSec`, rounded **up**, when each handles `perReplica` per second.

```go
ReplicasNeeded(1000, 300) // => 4
ReplicasNeeded(900, 300)  // => 3
ReplicasNeeded(1, 300)    // => 1
ReplicasNeeded(0, 300)    // => 0
```

The starter's `return` line divides, which rounds down and leaves you short.
`perReplica` is never 0.

**Think it through:**
- In the ceiling trick, which parameter plays `a` and which plays `b`?

### 3. Work out the CPU percentage

**Goal:** make `CPUPercent` return used divided by the limit, times 100.

```go
CPUPercent(250, 1000)  // => 25
CPUPercent(125, 1000)  // => 12.5
CPUPercent(1500, 1000) // => 150
```

Kubernetes measures CPU in *millicores*: 1000 millicores is one full CPU.
Above 100 means the container wants more than its limit and is slowed down.
`limitMillicores` is never 0.

**Think it through:**
- These are `float64`, so `/` keeps the fraction. Does the order of `/` and `*` matter here?

## Learn more

- [A Tour of Go: Basic types](https://go.dev/tour/basics/11)
- [Go spec: Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators)
- [Kubernetes: CPU resource units](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#meaning-of-cpu)
