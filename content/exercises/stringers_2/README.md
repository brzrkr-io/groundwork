# Stringers: pod phases and uptime

A pod dashboard that says `Failed` for a job that actually succeeded wakes
people up for nothing. You'll fix a broken `String` method, pick the right
formatting verb, and write a `String` method that does some arithmetic.

**New here:** lookup slices for enum names, splitting seconds with `/` and
`%`, `switch` with no value.

## Learn

### Recap

As in *Stringers*: a type with a `String() string` method satisfies
`fmt.Stringer`. `fmt` calls it for `%v`, `%s`, `%q`, `Print`, `Println` and
`Sprint`. The `%d` verb skips `String` and prints the raw number. Inside
`String`, convert the receiver (`int(p)`) before formatting it.

### A lookup table of names

Instead of a long `switch`, an enum can keep its names in a slice. Each
constant's number is the index of its name:

```go
type Color int

const (
	Red   Color = iota // 0
	Green              // 1
	Blue               // 2
)

var colorNames = []string{"Red", "Green", "Blue"}
//                         [0]    [1]      [2]

func (c Color) String() string {
	if c < 0 || int(c) >= len(colorNames) {
		return fmt.Sprintf("Color(%d)", int(c))
	}
	return colorNames[c]
}
```

The `if` guards against numbers outside the slice, which would crash the
program. The catch: the slice must list the names in **exactly** the same
order as the constants, with none missing.

> **Why is it written like this?** Go lets you index a slice with any
> integer type, including `Color`. A slice instead of a `switch` is the
> author's choice: shorter, but easy to get out of order.

### Splitting seconds into units

Integer division `/` gives how many whole units fit. The remainder operator
`%` gives what's left over:

```go
secs := 130
secs / 60 // 2   (whole minutes)
secs % 60 // 10  (seconds left over)
```

For bigger units, chain them: `secs % 3600` drops the whole hours, then
`/ 60` counts the whole minutes left.

> **Why is it written like this?** Go requires both sides of `/` to be
> integers for whole-number division; `7 / 2` is `3`, not `3.5`. `%` and `/`
> have the same precedence and run left to right.

### A switch with no value

`switch` can have nothing after it. Then each `case` is a `bool` condition,
and the first true one runs:

```go
switch {
case temp < 0:
	return "freezing"
case temp < 20:
	return "cool"
default:
	return "warm"
}
```

Order matters: put the smallest limit first, so each case only sees values
the earlier cases didn't take.

## Your task

You maintain the CLI that prints pod phases and uptimes for your cluster.

### 1. Fix the pod phase names

**Goal:** fix the `phaseNames` slice so every `Phase` prints its own name.

- The five constants, in order: `Pending`, `Running`, `Succeeded`, `Failed`,
  `Unknown`.
- Right now finished jobs show up as `Failed`.

```go
Succeeded.String() // => "Succeeded"
Failed.String()    // => "Failed"
Unknown.String()   // => "Unknown"
Phase(9).String()  // => "Phase(9)"
```

**Think it through:**
- Count the constants, then count the names in `phaseNames`.
- Line them up by index: which name sits at index 2?

### 2. Fix the status line

**Goal:** `PodStatusLine` must print the phase by name, not as a number.

```go
PodStatusLine("api-7f9c", Running)
// => "pod api-7f9c is Running"   (not "pod api-7f9c is 1")
```

**Think it through:**
- Which verb in the format string formats `phase`?
- Check the Recap: which verbs call `String`?

### 3. Print uptime like a human

**Goal:** finish `Uptime.String` so it shows the largest unit that fits,
then the next smaller unit.

An `Uptime` is a number of seconds. The constants `Minute` (60), `Hour`
(3600) and `Day` (86400) are already declared.

| Uptime | Output | Example |
|---|---|---|
| less than `Minute` | seconds, then `s` | `Uptime(45)` → `"45s"` |
| less than `Hour` | minutes `m`, a space, leftover seconds `s` | `Uptime(725)` → `"12m 5s"` |
| less than `Day` | hours `h`, a space, leftover minutes `m` | `Uptime(10859)` → `"3h 0m"` |
| `Day` or more | days `d`, a space, leftover hours `h` | `Uptime(187200)` → `"2d 4h"` |

- The second unit is always shown, even when it's 0.
- Anything smaller than the second unit is dropped.
- There's no unit bigger than days.

```go
Uptime(0).String()            // => "0s"
Uptime(60).String()           // => "1m 0s"
Uptime(5430).String()         // => "1h 30m"
fmt.Sprint(Uptime(400 * Day)) // => "400d 0h"
```

**Think it through:**
- Which row of the table does each example fall in? Each row is one case.
- For each row: which constant do you divide by for the big unit, and how do
  you get the leftover small unit?
- Convert `u` to a plain `int` first (see the Recap).

## Learn more

- [fmt.Stringer](https://pkg.go.dev/fmt#Stringer)
- [Kubernetes: Pod phase](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-phase)
- [Go spec: Index expressions](https://go.dev/ref/spec#Index_expressions)
