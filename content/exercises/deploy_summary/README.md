# Practice: deploy summary

Time to put the whole Basics module together. You'll write the helpers behind
a deploy bot's summary message, the kind that lands in a team chat channel
every time something ships:

```text
api v1.4.2 -> production: 3 replicas x 256 MiB = 768 MiB
```

## Learn

This exercise uses only what you've already learned:

| You need | From |
|---|---|
| parameters, `:=`, `+` on strings | *Variables* |
| `int`, `float64`, `*`, `/`, the ceiling trick | *Numbers and arithmetic* |
| `strings.TrimSpace`, `strings.ToLower` | *Strings* |
| `fmt.Sprintf`, `%s`, `%d`, `%.2f` | *Formatting with fmt* |
| `float64(...)` | *Type conversion* |

The file starts with no imports. Add `import "strings"` and `import "fmt"`
when you need them. You can group several imports like a `const` block:

```go
import (
	"fmt"
	"strings"
)
```

## Your task

Write the four helpers. Later ones may call earlier ones.

### 1. Name the release

People type service names with odd capitals and stray spaces. Write
`Release` so it trims the whitespace from both ends of the service name,
lower-cases it, then adds a space and the version (unchanged):

```go
Release(" API ", "v1.4.2")    // => "api v1.4.2"
Release("Billing", "v2.0.0")  // => "billing v2.0.0"
Release("worker", "v0.9.1")   // => "worker v0.9.1"
```

### 2. Size the deploy

Each replica serves `rpsPerReplica` requests per second. Write
`ReplicasNeeded` so it returns how many replicas cover `peakRPS`, rounded
**up**:

```go
ReplicasNeeded(1000, 300) // => 4
ReplicasNeeded(900, 300)  // => 3
ReplicasNeeded(50, 300)   // => 1
```

`rpsPerReplica` is never 0.

### 3. Write the summary line

Write `Summary` so it returns:

```text
<release> -> <env>: <replicas> replicas x <memoryMiB> MiB = <total> MiB
```

`<release>` is the result of `Release(service, version)`, and `<total>` is
`replicas` times `memoryMiB`.

```go
Summary("api", "v1.4.2", "production", 3, 256)
// => "api v1.4.2 -> production: 3 replicas x 256 MiB = 768 MiB"

Summary(" Worker ", "v0.9.1", "staging", 2, 512)
// => "worker v0.9.1 -> staging: 2 replicas x 512 MiB = 1024 MiB"
```

### 4. Estimate the monthly cost

Memory is billed per GiB per month, and 1 GiB is 1024 MiB. Write
`MonthlyCost` so it:

1. works out the total memory in MiB (`replicas` times `memoryMiB`),
2. converts it to GiB as a `float64` (divide by 1024, keeping the fraction),
3. multiplies by `dollarsPerGiB`,
4. returns `$`, the cost with exactly 2 decimals, then `/month`.

```go
MonthlyCost(3, 512, 4)    // => "$6.00/month"   (1.5 GiB x $4)
MonthlyCost(4, 256, 3.5)  // => "$3.50/month"   (1 GiB x $3.50)
MonthlyCost(10, 768, 2.2) // => "$16.50/month"  (7.5 GiB x $2.20)
```

## Learn more

- [Go by Example: String Formatting](https://gobyexample.com/string-formatting)
- [Package strings](https://pkg.go.dev/strings)
- [Kubernetes: Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
