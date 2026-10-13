## Task 1: count the names

There are five constants but only four names in `phaseNames`. Line them up:
index 0 is `Pending`, index 1 is `Running`, and so on. Which one is missing?

## Task 2: pick the verb

`%d` prints a `Phase` as its number. Which verbs make `fmt` call `String`?
Look at the table in *Stringers* or the recap in this README.

## Task 3: one case per unit

Convert once with `s := int(u)`. A `switch` with no value after it runs the
first `case` whose condition is true, so check the smallest unit first:

```go
switch {
case s < Minute:
	// seconds only
case s < Hour:
	// minutes and leftover seconds
// ...
}
```

## Almost the answer

```go
var phaseNames = []string{"Pending", "Running", "Succeeded", "Failed", "Unknown"}
```

```go
return fmt.Sprintf("pod %s is %v", name, phase)
```

```go
case s < Hour:
	return fmt.Sprintf("%dm %ds", s/Minute, s%Minute)
case s < Day:
	return fmt.Sprintf("%dh %dm", s/Hour, s%Hour/Minute)
```

Write the days case the same way, using `default:`.
