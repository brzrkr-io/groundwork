## Task 1: Write the status line

First add `import "fmt"`. Then use `fmt.Sprintf`: write the whole sentence as one format string, with a verb where each value goes.

### Closer

Fixed text stays in the format string; the parameters come after it, in order. `%s` is for a string and `%d` for an int: `fmt.Sprintf("%s: ...", service, replicas)`.

### Answer

```go
return fmt.Sprintf("%s: %d replicas ready", service, replicas)
```

## Task 2: Report rollout progress

You need three things in the format string: a verb that adds quotes, a verb for a float with one decimal, and a literal percent sign.

### Closer

`%q` quotes a string for you, `%.1f` gives one digit after the point, and `%%` prints a single `%`. The format starts `"rollout %q ..."`.

### Answer

```go
return fmt.Sprintf("rollout %q %.1f%% complete", service, percent)
```

## Task 3: Line up a table row

Put a number between `%` and the verb to set a width. A minus sign makes it left-aligned.

### Closer

`%-10s` pads a string on the right to 10 characters; `%5d` pads a number on the left to 5. The ` | ` separators are fixed text inside the format string.

### Answer

```go
return fmt.Sprintf("%-10s | %5d | %s", service, replicas, status)
```
