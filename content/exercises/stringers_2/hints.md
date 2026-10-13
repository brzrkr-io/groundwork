## Task 1: Fix the pod phase names

There are five constants but only four names in `phaseNames`. Line them up:
index 0 is `Pending`, index 1 is `Running`, and so on. Which one is missing?

### Closer

`Succeeded` is constant number 2, but index 2 of the slice holds `"Failed"`.
Add `"Succeeded"` at the right position so every name sits at its constant's
number.

### Answer

```go
var phaseNames = []string{"Pending", "Running", "Succeeded", "Failed", "Unknown"}
```

## Task 2: Fix the status line

`%d` prints a `Phase` as its raw number. Which verbs make `fmt` call `String`?
Look at the recap in this README.

### Closer

`%v` and `%s` both call `String`. Swap the verb that formats `phase`.

### Answer

```go
return fmt.Sprintf("pod %s is %v", name, phase)
```

## Task 3: Print uptime like a human

Convert once with `s := int(u)`, then pick the unit by comparing `s` with
`Minute`, `Hour` and `Day`. Check the smallest unit first, so the first match
is the right one.

### Closer

A `switch` with no value after it runs the first `case` whose condition is
true. Use `/` for whole units and `%` for what is left:

```go
switch {
case s < Minute:
	return fmt.Sprintf("%ds", s)
case s < Hour:
	return fmt.Sprintf("%dm %ds", s/Minute, s%Minute)
// ...
}
```

### Answer

```go
s := int(u)
switch {
case s < Minute:
	return fmt.Sprintf("%ds", s)
case s < Hour:
	return fmt.Sprintf("%dm %ds", s/Minute, s%Minute)
case s < Day:
	return fmt.Sprintf("%dh %dm", s/Hour, s%Hour/Minute)
default:
	return fmt.Sprintf("%dd %dh", s/Day, s%Day/Hour)
}
```

`s%Hour/Minute` first drops the whole hours, then counts the whole minutes in what is left.
