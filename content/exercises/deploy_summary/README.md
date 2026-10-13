# Practice: deploy summary

Time to put the whole Basics module together. You'll write the helpers behind
a deploy bot's summary message, the kind that lands in a team chat channel
every time something ships:

```text
api v1.4.2 -> production: 3 replicas x 256 MiB = 768 MiB
```

**New here:** grouped imports; otherwise practice.

## Learn

This exercise uses only what you've already learned:

| You need | From |
|---|---|
| parameters, `:=`, `+` on strings | *Variables* |
| `int`, `float64`, `*`, `/`, the ceiling trick | *Numbers and arithmetic* |
| `strings.TrimSpace`, `strings.ToLower` | *Strings* |
| `fmt.Sprintf`, `%s`, `%d`, `%.2f` | *Formatting with fmt* |
| `float64(...)`, calling your own function | *Type conversion* |

### Grouped imports

The file starts with no imports. You'll need `strings` and `fmt`. Several
imports can be grouped like a `const` block:

```go
import (
	"os"
	"time"
)
```

> **Why is it written like this?** **Go requires** each path in quotes, and
> every import to be used. **Convention:** one `import ( ... )` block, with
> paths sorted A–Z (gofmt sorts them for you). **Your choice:** one block or
> separate `import` lines; both compile.

### Expressions as arguments

A `Sprintf` value doesn't have to be a plain variable. Any expression of the
right type works: a calculation, or a call to another function.

```go
fmt.Sprintf("%d jobs, %s", workers*perWorker, strings.ToUpper(state))
```

## Your task

Write the four helpers. Later ones may call earlier ones. Add the imports on
line 4, where `TODO(1)` says.

### 1. Name the release

**Goal:** make `Release` trim the whitespace from both ends of `service`,
lower-case it, then add a space and `version` (unchanged).

```go
Release(" API ", "v1.4.2")   // => "api v1.4.2"
Release("Billing", "v2.0.0") // => "billing v2.0.0"
Release("worker", "v0.9.1")  // => "worker v0.9.1"
```

**Think it through:**
- Which two `strings` functions clean the name, and which runs first?
- The space between name and version is fixed text.

### 2. Size the deploy

**Goal:** make `ReplicasNeeded` return how many replicas cover `peakRPS`,
rounded **up**, when each serves `rpsPerReplica`.

```go
ReplicasNeeded(1000, 300) // => 4
ReplicasNeeded(900, 300)  // => 3
ReplicasNeeded(50, 300)   // => 1
```

`rpsPerReplica` is never 0.

**Think it through:**
- This is the ceiling trick from *Numbers and arithmetic*. Which parameter is `a`, which `b`?

### 3. Write the summary line

**Goal:** make `Summary` return this line:

```text
<release> -> <env>: <replicas> replicas x <memoryMiB> MiB = <total> MiB
```

`<release>` is what your `Release` function gives for this `service` and
`version`. `<total>` is `replicas` times `memoryMiB`. Everything else is fixed
text.

```go
Summary("api", "v1.4.2", "production", 3, 256)
// => "api v1.4.2 -> production: 3 replicas x 256 MiB = 768 MiB"

Summary(" Worker ", "v0.9.1", "staging", 2, 512)
// => "worker v0.9.1 -> staging: 2 replicas x 512 MiB = 1024 MiB"
```

**Think it through:**
- Five values, so five verbs. Which are strings, which are ints?
- Two values aren't parameters: one is a call, one a calculation.

### 4. Estimate the monthly cost

**Goal:** make `MonthlyCost` return `$`, the monthly cost with exactly 2
decimals, then `/month`.

Memory is billed per GiB per month, and 1 GiB is 1024 MiB. The steps:

1. total memory in MiB: `replicas` times `memoryMiB`,
2. convert it to GiB as a `float64` (divide by 1024, keeping the fraction),
3. multiply by `dollarsPerGiB`,
4. format it.

```go
MonthlyCost(3, 512, 4)    // => "$6.00/month"   (1.5 GiB x $4)
MonthlyCost(4, 256, 3.5)  // => "$3.50/month"   (1 GiB x $3.50)
MonthlyCost(10, 768, 2.2) // => "$16.50/month"  (7.5 GiB x $2.20)
```

**Think it through:**
- Step 1 is `int`; step 3 needs `float64`. Where does the conversion go?
- `$` and `/month` are fixed text around one verb.

## Learn more

- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
- [Package strings](https://pkg.go.dev/strings)
- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
