## Task 1: Build a command line

Write `bin` first, then each argument with a space before it. Before writing an argument, decide whether it needs double quotes: it does when it is empty or contains a space.

### Closer

Add `import "strings"`. Use a `strings.Builder` and loop over `args`. The check is `a == "" || strings.Contains(a, " ")`, and a raw string `` `"` `` is a double quote, so wrapping is `` `"` + a + `"` ``.

### Answer

```go
var b strings.Builder
b.WriteString(bin)
for _, a := range args {
	b.WriteString(" ")
	if a == "" || strings.Contains(a, " ") {
		a = `"` + a + `"`
	}
	b.WriteString(a)
}
return b.String()
```

Each argument gets one space in front, so there's no extra space at the start or end.

## Task 2: Target a namespace

Let `Command` do the work: call it with `"kubectl"` and the right arguments. You can't write `Command("kubectl", "-n", namespace, args...)`, so build one slice of arguments first.

### Closer

Start with `var all []string`. Only when `namespace != ""`, do `all = append(all, "-n", namespace)`. Then append `args...` and pass `all...` to `Command`.

### Answer

```go
var all []string
if namespace != "" {
	all = append(all, "-n", namespace)
}
all = append(all, args...)
return Command("kubectl", all...)
```

Appending to a `nil` slice works fine, so the empty-namespace case needs no special handling.

## Task 3: Chain the steps

This is the same pattern as `JoinPath` in *Variadic functions*, but with `" && "` instead of `"/"`. Skip empty commands and only put the separator between commands.

### Closer

Use `continue` for `c == ""`. Before adding a command, add `" && "` only if the result so far is not empty.

### Answer

```go
script := ""
for _, c := range commands {
	if c == "" {
		continue
	}
	if script != "" {
		script += " && "
	}
	script += c
}
return script
```

The separator only goes in when there's already a command before it.
