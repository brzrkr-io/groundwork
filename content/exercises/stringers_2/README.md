# Stringers: pod phases and uptime

A pod dashboard that says `Failed` for a job that actually succeeded wakes
people up for nothing. This time you fix a broken `String` method, choose the
right formatting verb, and write a `String` method that does some arithmetic.

## Learn

### Recap

A type with a `String() string` method satisfies `fmt.Stringer`, and `fmt`
calls that method for `%v`, `%s`, `%q`, `Print`, `Println` and `Sprint`. The
`%d` verb skips `String` and prints the raw number. Inside `String`, convert
the receiver (`int(p)`) before you format it, as in *Stringers*.

### A lookup table of names

Instead of a long `switch`, an enum can keep its names in a slice. The
constant's number is the index of its name:

```go
type Color int

const (
	Red   Color = iota // 0
	Green              // 1
	Blue               // 2
)

var colorNames = []string{"Red", "Green", "Blue"}

func (c Color) String() string {
	if c < 0 || int(c) >= len(colorNames) {
		return fmt.Sprintf("Color(%d)", int(c))
	}
	return colorNames[c]
}
```

You can index a slice with any integer type, including `Color`. The `if`
guards against numbers outside the slice, which would otherwise crash the
program. The catch: the slice must list the names in **exactly** the same
order as the constants, with none missing.

### Splitting seconds into units

`Uptime` below is a number of seconds. Integer division `/` tells you how
many whole units fit, and the remainder operator `%` tells you what is left
over:

```go
secs := 725
secs / 60 // 12  (whole minutes)
secs % 60 // 5   (seconds left over)
```

The same works for bigger units: `secs / 3600` is whole hours, and
`secs % 3600 / 60` is the minutes left after taking out the hours.

## Your task

You maintain the CLI that prints pod phases and uptimes for your cluster.

### 1. Fix the pod phase names

`Phase` has five values: `Pending`, `Running`, `Succeeded`, `Failed` and
`Unknown`. Its `String` method uses a lookup slice, but finished jobs show up
as `Failed`. Fix `phaseNames` so every phase prints its own name:

```go
Succeeded.String() // => "Succeeded"
Failed.String()    // => "Failed"
Unknown.String()   // => "Unknown"
Phase(9).String()  // => "Phase(9)"
```

### 2. Fix the status line

`PodStatusLine` prints the phase as a number. Change it so the phase is
printed by name:

```go
PodStatusLine("api-7f9c", Running)
// => "pod api-7f9c is Running"   (not "pod api-7f9c is 1")
```

### 3. Print uptime like a human

Finish `Uptime.String`. An `Uptime` is a number of seconds. Show the largest
unit that fits, followed by the next smaller unit, and drop everything
smaller than that:

| Uptime | Output | Example |
|---|---|---|
| less than `Minute` (60) | seconds, then `s` | `Uptime(45)` → `"45s"` |
| less than `Hour` (3600) | minutes `m`, a space, leftover seconds `s` | `Uptime(725)` → `"12m 5s"` |
| less than `Day` (86400) | hours `h`, a space, leftover minutes `m` | `Uptime(10859)` → `"3h 0m"` |
| `Day` or more | days `d`, a space, leftover hours `h` | `Uptime(187200)` → `"2d 4h"` |

The leftover unit is always shown, even when it is 0, and there is no unit
bigger than days. More examples:

```go
Uptime(0).String()    // => "0s"
Uptime(60).String()   // => "1m 0s"
Uptime(5430).String() // => "1h 30m"
fmt.Sprint(Uptime(400 * Day)) // => "400d 0h"
```

## Learn more

- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [Kubernetes: Pod phase](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase)
- [Go spec: Index expressions](https://go.dev/ref/spec#Index_expressions)
